package app

import (
	"flag"
	"fmt"
)

type Flag interface {
	fmt.Stringer
	GetName() string
	GetValue() any
	Apply(*flag.FlagSet)
}

type StringFlag struct {
	Name  string
	Value string
	Usage string
}

type IntFlag struct {
	Name  string
	Value int
	Usage string
}

type BoolFlag struct {
	Name  string
	Value bool
	Usage string
}

func (f *StringFlag) Apply(fs *flag.FlagSet) {
	fs.StringVar(&f.Value, f.Name, f.Value, f.Usage)
}

func (f *StringFlag) String() string {
	return fmt.Sprintf("%s=%s", f.Name, f.Value)
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) GetValue() any {
	return f.Value
}

func (f *IntFlag) Apply(fs *flag.FlagSet) {
	fs.IntVar(&f.Value, f.Name, f.Value, f.Usage)
}

func (f *IntFlag) String() string {
	return fmt.Sprintf("%s=%d", f.Name, f.Value)
}

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) GetValue() any {
	return f.Value
}

func (f *BoolFlag) Apply(fs *flag.FlagSet) {
	fs.BoolVar(&f.Value, f.Name, f.Value, f.Usage)
}

func (f *BoolFlag) String() string {
	return fmt.Sprintf("%s=%t", f.Name, f.Value)
}

func (f *BoolFlag) GetName() string {
	return f.Name
}

func (f *BoolFlag) GetValue() any {
	return f.Value
}
