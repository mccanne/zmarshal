package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/brimdata/zed/zson"
)

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

func Make(which string) Thing {
	if which == "rose" {
		return &Plant{"red"}
	}
	if which == "ivy" {
		return &Plant{"green"}
	}
	if which == "flamingo" {
		return &Animal{"pink"}
	}
	return nil
}

func ex1() {
	rose := Make("rose")
	flamingo := Make("flamingo")
	m := zson.NewMarshaler()
	m.Decorate(zson.StyleSimple)
	roseZSON, _ := m.Marshal(rose)
	fmt.Println(roseZSON)
	flamingoZSON, _ := m.Marshal(flamingo)
	fmt.Println(flamingoZSON)
}

func ex2() {
	f := Make("flamingo")
	m := zson.NewMarshaler()
	m.Decorate(zson.StyleSimple)
	flamingoZSON, _ := m.Marshal(f)

	u := zson.NewUnmarshaler()
	u.Bind(Animal{}, Plant{})
	var flamingo Thing
	err := u.Unmarshal(flamingoZSON, &flamingo)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("The flamingo is " + flamingo.Color())
	}
}

func ex3() {
	f := Make("flamingo")
	m := zson.NewMarshaler()
	m.Decorate(zson.StyleSimple)
	flamingoZSON, _ := m.Marshal(f)

	u := zson.NewUnmarshaler()
	u.Bind(Animal{}, Plant{})
	var flamingo Thing
	u.Unmarshal(flamingoZSON, &flamingo)
	_, ok := flamingo.(*Animal)
	fmt.Printf("The flamingo is an Animal? %t\n", ok)
}

func ex4() {
	rose := Make("rose")
	flamingo := Make("flamingo")
	m := zson.NewMarshaler()
	m.Decorate(zson.StylePackage)
	roseZSON, _ := m.Marshal(rose)
	fmt.Println(roseZSON)
	flamingoZSON, _ := m.Marshal(flamingo)
	fmt.Println(flamingoZSON)
}

func ex5() {
	rose := Make("rose")
	flamingo := Make("flamingo")
	m := zson.NewMarshaler()
	m.NamedBindings([]zson.Binding{{"Plant.v0", Plant{}}, {"Animal.v0", Animal{}}})
	roseZSON, _ := m.Marshal(rose)
	fmt.Println(roseZSON)
	flamingoZSON, _ := m.Marshal(flamingo)
	fmt.Println(flamingoZSON)
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	ex, err := strconv.Atoi(os.Args[1])
	if err != nil {
		usage()
	}
	switch ex {
	case 1:
		ex1()
	case 2:
		ex2()
	case 3:
		ex3()
	case 4:
		ex4()
	case 5:
		ex5()
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: zmarshal example-number")
	os.Exit(1)
}
