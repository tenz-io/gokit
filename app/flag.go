package app

import (
	"flag"
)

type Flag interface {
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

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) GetValue() any {
	return f.Value
}

func (f *IntFlag) Apply(fs *flag.FlagSet) {
	fs.IntVar(&f.Value, f.Name, f.Value, f.Usage)
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

func (f *BoolFlag) GetName() string {
	return f.Name
}

func (f *BoolFlag) GetValue() any {
	return f.Value
}
