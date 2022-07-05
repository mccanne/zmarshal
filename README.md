# Unmarshal Go interface values with ease

> TL;DR You can finally unmarshal Go interface values with ease using
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
very same data back into the very same interface value.

Why is this?

Let's look at a concrete example.  We'll follow the patterns used in
[Greg Trowbridge's article](http://gregtrowbridge.com/golang-json-serialization-with-interfaces/)
on this topic, where he first creates a `Plant` type and an `Animal` type, which
both implement a `Thing` interface:
```
type Thing interface {  
    Color() string
}

type Plant struct {  
    MyColor string
}

func (p *Plant) Color() string { return p.MyColor }

type Animal struct {  
    MyColor string
}

func (a *Animal) Color() string { return a.MyColor }

```
With this pattern, let's make a Plant and marshal it into JSON:
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
[pre-loaded into the Go Playground](https://go.dev/play/p/SEWJut0mCz-).
Just hit the Run button.

[ex1](https://go.dev/play/p/dYFPNbUANUV)

## Marshaling Interfaces

Okay, we successively marshaled a Go struct, but what about an interface value?
Fortunately, the marshaling logic here will work just fine for a Thing type.
Suppose we get an interface value from somewhere like this:
```
func Make(which, color string) Thing {
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
And now, if we marshal a Thing, like so,
```
flamingo := Make("animal", "pink")
flamingoJSON, _ := json.Marshal(flamingo)
fmt.Println(string(flamingoJSON))
```
we'll get the following output
[(try it)](https://go.dev/play/p/aioSK48Whoi):
```
{"MyColor":"pink"}
```
Perfect.  Go's json marshaler followed the interface value to its implementation
and output exactly what we wanted.  Now, let's try to Unmarshal the JSON back
into an interface type, e.g.,
[(try it)](https://go.dev/play/p/_rjk7Uhi4I6):
```
	var thing Thing
	err := json.Unmarshal(flamingoJSON, &thing)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(thing.Color())
	}
```
Oh no, we get an error that looks like this:
```
json: cannot unmarshal object into Go value of type main.Thing
```
Why can't Go's json package unmarshal this object?  That encoding is exactly
what the Marshal function produced when we marshaled the flamingo object
in the first place.

What gives?

Trowbridge boils this down to a very simple observation: what if
we looked at the two JSON serializations _from Go's perspective?_

To do so, here is a snippet to serialize a flamingo and a rose
[(try it)](https://go.dev/play/p/6SL21fiXuom):
```
rose := Make("plant", "red")
roseJSON, _ := json.Marshal(rose)
fmt.Println(string(roseJSON))
flamingo := Make("animal", "pink")
flamingoJSON, _ := json.Marshal(flamingo)
fmt.Println(string(flamingoJSON))
rose := Make("rose")
```
And, we get this output:
```
{"MyColor":"red"}
{"MyColor":"pink"}
```
Now the problem starts to make sense!

The JSON output here is exactly the same for the
both the `Plant` and the `Animal`.  How is Go supposed to figure out which is which?

You might say, in this case, the compiler could look at the color strings in the
value and figure out which struct to use because a rose is red and a flamingo is pink.
Okay sure, but in general, there just isn't enough information here, as we could
also have a red cardinal and the cardinal and the rose would have
exactly the same JSON in this example.

The fundamental issue here is that neither the _planted-ness_ of the rose nor
the _animal-ness_ of the flamingo made it into the JSON output.  Alas, you say,
the solution is just a small matter of programming: add a plant/animal type field to the
JSON output and you're golden.  In fact, Go's `json` package makes this approach
all quite feasible with its custom marshaler methods.  Trowbridge walks you through
how to do this, and after a number of non-obvious steps (especially if
you're new to Go) and a hundred or so lines of code, he declares victory
at the end of the article: "YOU MADE IT!"

Is this the best we got?  Surely there's got to be a better way.

## Enter ZSON

What if there were a data format like JSON but it could reflect the
Go types into its serialized representation so the _planted-ness_
and _animal-ness_ from our example above could be handled automatically?

It turns out there is new type of data called
[super-structured data](https://zed.brimdata.io/docs/formats/#2-zed-a-super-structured-pattern)
that can carry the information needed to solve out problem here.

We won't go into all the gory details of super-structured data
but suffice it to say that ZSON provides a comprehensive type system
that can reliably represent any serializable Go type and includes type definitions and first-class
type values so it can carry the type names of Go values into its
serialized form.

We'll use the ZSON form of super-structured data.  ZSON is a superset of JSON
so it will look familiar but it carries the full power of the super-structured
data model.

Armed with ZSON, we can serialize the flamingo and rose with the
super-structured type information
[(try it)](https://go.dev/play/p/ZhC-vr7xm3X):
```
rose := Make("plant", "red")
flamingo := Make("animal", "pink")
m := zson.NewMarshaler()
m.Decorate(zson.StyleSimple)
roseZSON, _ := m.Marshal(rose)
fmt.Println(roseZSON)
flamingoZSON, _ := m.Marshal(flamingo)
fmt.Println(flamingoZSON)
```
And, we get this output:
```
{MyColor:"red"}(=Plant)
{MyColor:"pink"}(=Animal)
```
As you can see, the _planted-ness_ and _animal-ness_ of the `Thing` is
noted in the output!

The parenthesized strings at the end of each line are are called ZSON
"type decorators".  ZSON has a full-fledged type system and these decorators
may be embedded throughout complex and highly nested ZSON values to provide
precise type semantics.  Mind you, these type names look like Go-specific
type values but there is nothing language-specific in the type name.  It can
be any string, but it just so happens the Go marshaler chooses ZSON type names
to match the Go types being serialized.

Given the type information in the ZSON output, we should be able to unmarshal the
ZSON back into an interface value, right?  There's one little twist.
Because Go doesn't have a way to convert the name of type to a value of that
type, you need to help out the ZSON unmarshaler by giving it a list
of values that might be referenced in the ZSON using the `Bind()`
method on the unmarshaler.  Here's how this works
[(try it)](https://go.dev/play/p/aZCw3R7W3lp):
```
u := zson.NewUnmarshaler()
u.Bind(Animal{}, Plant{})
var thing Thing
if err := u.Unmarshal(flamingoZSON, &thing); err != nil {
	fmt.Println(err)
} else {
	fmt.Println("The flamingo is " + thing.Color())
}
if err := u.Unmarshal(roseZSON, &thing); err != nil {
	fmt.Println(err)
} else {
	fmt.Println("The rose is " + thing.Color())
}
```
If you run this, you will see the serialized ZSON values
are successfully marshaled into the interface variable with the correct
underlying concrete types.  The output here is:
```
The flamingo is pink
The rose is red
```
Just for good measure, you can see here that
the type of concrete value is in fact correct
([try it](https://go.dev/play/p/nq6k6n3EoNs)):
```
	_, ok := flamingo.(*Animal)
	fmt.Printf("Is the flamingo an Animal? %t\n", ok)
```
and the output is
```
Is the flamingo an Animal? true
```

### Custom Type Names

You probably noticed in these examples, the ZSON marshaling used the exact
same type names as the Go program.  This can create name conflicts since
the same type name may appear in different Go packages (e.g., io.Writer and
bufio.Writer).

To cope with this, the ZSON marshaler let's you specify more detailed types by
providing zson.TypeStyle to the marshaler's Decorate method.  You can use
package names with `zson.StylePackage`, e.g., `zmarshal 4` will produce this
output:
```
{MyColor: "red"}(=main.Plant)
{MyColor: "pink"}(=main.Animal)
```
Type names can also be extended to include the full important path using
`zson.StyleFull` and even include version numbers in the type path to provide
a mechanism for versioning the "schema" of these serialized messages.
For example, `zmarshal 5` utilizes the `NamedBindings` method on the marshaler
to establish a binding between the chosen type name and the Go data structure
[try it](https://go.dev/play/p/QhyaT_lTM5B):
```
	rose := Make("plant", "red")
	flamingo := Make("animal", "pink")
	m := zson.NewMarshaler()
	m.NamedBindings([]zson.Binding{{"Plant.v0", Plant{}}, {"Animal.v0", Animal{}}})
	roseZSON, _ := m.Marshal(rose)
	fmt.Println(roseZSON)
	flamingoZSON, _ := m.Marshal(flamingo)
	fmt.Println(flamingoZSON)
```
and produces this output:
```
{MyColor:"red"}(=Plant.v0)
{MyColor:"pink"}(=Animal.v0)
```
Then down the road, as you enhanced the `Animal` and `Plant` types,
you could imagine unmarshaling multiple versions of the `Thing`,
with different ZSON version numbers,
format into different concrete types all behind a single Go interface value.

## Wrapping Up

* streams of marshaled values (e.g., logs as opposed to config or app state)
* if you have streams of data, you can search and analyze them with `zq`
