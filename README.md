# zgom - Go marshaler for ZSON

> TL;DR You can finally unmarshal Go interface types with ease using
> [ZSON](https://github.com/brimsec/zq/blob/master/zng/docs/zson.md).
> ZSON is a new dialect of JSON, which embeds a comprehensive
> type system in a syntactically friendly fashion.
> When marshaling an interface value into ZSON, the
> type name of the interface's underlying implementation is reflected
> into ZSON as a ZSON
> [first-class type name](https://github.com/brimsec/zq/blob/master/zng/docs/zson.md#25-type-definitions).

Have you ever gotten frustrated unmarshaling JSON into a Go interface value?
Turns out you're not the only one!

If you know what I'm talking about,
[you can cut to the chase](#enter-zson), but if you are a mortal being
like most of us, and you find Go interfaces a challenge to marshal, please read on.

## The Problem

While the Go json.Marshal function does a wonderful job marshaling interface
values into JSON, there is an odd asymmetry when it comes to unmarshaling the
very same data back into they very same interface value.

Why is this?

Let's look at a concrete example.  We'll follow the patterns used in
[Greg Trowbridge's article](http://gregtrowbridge.com/golang-json-serialization-with-interfaces/)
on this topic, where he first creates a `Plant` type and an `Animal` type, which
both implement a `ColoredThing` interface:
```
type ColoredThing interface {  
    Color() string
}

type Plant struct {  
    MyColor string
}

func (p *Plant) Color() string { return p.MyColor }

type Animal struct {  
    MyColor string
}

func (a *Animal) Color() string { return .MyColor }

```
With this pattern, let's make a Plant and marshal it to JSON:
```
p := Plant{MyColor: "green"}
byteSlice, _ := json.Marshal(p)
fmt.Println(string(byteSlice))
```
this is, of course, prints out
```
{"MyColor":"green"}
```
You can try on this example
[pre-loaded into the Go Playground](https://play.golang.org/p/9tBwzh2WTZm).
Just hit the Run button.

## Marshaling Interfaces

Okay, we successively marshaled a Go struct, but what about an inteface.
Fortunately, the marshaling logic here will work just fine for a ColoredThing.
Suppose we get an interface value from somewhere like this:
```
func Make(which, color string) ColoredThing {
        switch which {
        case "plant":
                return &Plant{color}
        case "animal":
                return &Animal{color}
        default:
                return nil
        }
}
```
And now, if we marshal a ColoredThing, like so,
```
flamingo := Make("animal", "pink")
flamingoJSON, _ := json.Marshal(flamingo)
fmt.Println(string(flamingoJSON))
```
we'll get the following output
([try it](https://play.golang.org/p/sodGUg71_58)):
```
{"MyColor":"pink"}
```
Perfect.  Go's json marshaler followed the interface value to its implementation
and output exactly what we wanted.  Now, let's try to Unmarshal this back
into an interface type, e.g.,
([try it](https://play.golang.org/p/Rt1vlEZh2lO)):
```
	var flamingo ColoredThing
	err := json.Unmarshal(flamingoJSON, &flamingo)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(flamingo.Color())
	}
```
Oh no, we get an error that looks like this:
```
json: cannot unmarshal object into Go value of type main.ColoredThing
```
Why can't the json packet unmarshal this object?  That encoding is exactly
what the Marshal function produced when we marshalled the flamingo object
in the first place.

What gives?

Trowbridge adroitly boils this all down a very simple technique: what if
we looked at the different JSON serializations here, and look at the problem
of deserializing the JSON into an interface value _from Go's perspective_.

To do so, here is a snippet to serialize a flamingo and a rose
([try it]([https://play.golang.org/p/WgEUuC_XcQW)):
```
rose := Make("rose")
roseJSON, _ := json.Marshal(rose)
fmt.Println(string(roseJSON))
flamingo := Make("flamingo")
flamingoJSON, _ = json.Marshal(flamingo)
fmt.Println(string(flamingoJSON))
```
And, we get this output:
```
{"MyColor":"red"}
{"MyColor":"pink"}
```
Now the problem starts to make sense!  The JSON output here is exactly the same for the
both the Plant and the Animal.  How is Go supposed to figure out which is which?
You might say, in this case, the compiler could look at the color strings in the
value and figure out which struct to use because a rose is red and a flamingo is pink.
Okay sure, but in general, there just isn't enough information here, as we could also have
a red cardinal, in which case, the cardinal and the rose would have
indistinguishable serializations.

The fundamental issue here is that neither the _planted-ness_ of the rose nor
the _animal-ness_ of the flamingo made it into the JSON output.  Alas, you say,
that's just a small matter of programming: add a plant/animal type field to the
JSON output and you're golden.  In fact, Go's json package makes this approach
all quite feasible with its custom marshaler methods.  Trowbridge walks you through
how to do this, and after a number of fairly mind-bending steps (especially if
you're new to Go) and a hundred or so lines of code, he declares victory
at the end of the article: "YOU MADE IT!"

Is this the best we got?  Surely there's got to be a better way.

## Enter ZSON

What if there were a data format like JSON but it could reflect the
Go types into its serialized representation so the _planted-ness_
and _animal-ness_ could be handled automatically?

It turns out there is brand new format called ZSON that does just this.
We won't go into all the gory details of ZSON but suffice it to say that
ZSON provides a comprehensive type system that can reliably represent
any serializable Go type and includes type definitions and first-class
type values so it can carry the type names of Go values into its
serialized form.

Armed with ZSON, let's again serialize a flamingo and a rose
([try it]([https://TBD)):
```
rose := Make("rose")
flamingo := Make("flamingo")
m := zson.NewMarshaler()
m.Decorate(zson.TypeSimple)
roseZSON, _ := m.Marshal(rose)
fmt.Println(roseZSON)
flamingoZSON, _ = m.Marshal(flamingo)
fmt.Println(flamingoZSON)
```
And now, we get this magical output:
```
{MyColor:"red"} (=Plant)
{MyColor:"pink"} (=Animal)
```
As you can see, the _planted-ness_ and _animal-ness_ is preserved!

The parenthesized strings at the end of each line are are called ZSON
"type decorators".  ZSON has a full-fledged type system and these decorators
may be embedded throughout complex and highly nested ZSON values to provide
precise type semantics.  Mind you, these type names look like Go-specific
type values but there is nothing language-specific in the type name.  It can
be any string, but it just so happens the Go marshaler chooses ZSON type names
to match the Go types being serialized.

Given the type information in the ZSON output, we should be unmarshal the
ZSON back into an interface value, right?  There's one little twist.
Because Go doesn't have a way to convert the name of type to a value of that
type, you need to help out the ZSON unmarshaler by giving it a list
of values that might be referenced in the ZSON using the `Bindings()`
method on the unmarshaler.  Here's how this works
([try it]([https://TBD)):
```
u := zson.NewUnmarshaler()
u.Bindings(Animal{}, Plant{})
var flamingo ColoredThing
err := json.Unmarshal(flamingoJSON, &flamingo)
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(flamingo.Color())
}
```
And voila!  The concrete ZSON type was successfully marshaled into the
interface variable.  Just for good measure, you can see here that
the type of concrete value is in fact correct
([try it]([https://TBD)):
```
_, ok := flamingo.(*Animal)
fmt.Println(ok)
```

TBD:
* conflicting type names and `m.Decorate(zson.TypeFull)`
* streams of marshaled values (e.g., logs as opposed to config or app state)
* if you have streams of data, you can search and analyze them with `zq`
