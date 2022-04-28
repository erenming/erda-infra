// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	urlPackage      = protogen.GoImportPath("net/url")
	urlencPackage   = protogen.GoImportPath("github.com/erda-project/erda-infra/pkg/urlenc")
	stringsPackage  = protogen.GoImportPath("strings")
	structpbPackage = protogen.GoImportPath("google.golang.org/protobuf/types/known/structpb")
	base64Package   = protogen.GoImportPath("encoding/base64")
	jsonPackage     = protogen.GoImportPath("encoding/json")
	strconvPackage  = protogen.GoImportPath("strconv")
)

func generateFile(gen *protogen.Plugin, file *protogen.File) (*protogen.GeneratedFile, error) {
	filename := file.GeneratedFilenamePrefix + ".form.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by ", genName, ". DO NOT EDIT.")
	g.P("// Source: ", file.Desc.Path())
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the ", urlencPackage, " package it is being compiled against.")
	for _, message := range file.Messages {
		if strings.Contains(strings.ToUpper(message.Comments.Leading.String()), "+SKIP_GO-FORM") {
			continue
		}
		g.P("var _ ", urlencPackage.Ident("URLValuesUnmarshaler"), " = ", "(*", message.GoIdent.GoName, ")(nil)")
	}
	for _, message := range file.Messages {
		if strings.Contains(strings.ToUpper(message.Comments.Leading.String()), "+SKIP_GO-FORM") {
			continue
		}
		err := genMessage(gen, file, g, message)
		if err != nil {
			return g, err
		}
	}
	g.P()
	return g, nil
}

func genMessage(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) error {
	g.P()
	g.P("// ", message.GoIdent.GoName, " implement ", urlencPackage.Ident("URLValuesUnmarshaler"), ".")
	g.P("func (m *", message.GoIdent.GoName, ") UnmarshalURLValues(prefix string, values ", urlPackage.Ident("Values"), ") error {")
	params := createQueryParams(message.Fields)
	if len(params) > 0 {
		g.P("	for key, vals := range values {")
		g.P("		if len(vals) > 0 {")
		g.P("		switch prefix+key {")
		for _, param := range params {
			g.P("	case ", strconv.Quote(param.Name), ":")
			genQueryString(g, "m", strings.Split(param.Name, "."), param.Root.GoName, param.Root.Desc, param.Root.Message)
		}
		g.P("		}")
		g.P("		}")
		g.P("	}")
	}
	g.P("	return nil")
	g.P("}")
	return nil
}

func genQueryString(g *protogen.GeneratedFile, prefix string, names []string, goName string, desc protoreflect.FieldDescriptor, subMsg *protogen.Message) error {
	switch len(names) {
	case 0:
		return nil
	case 1:
		return genQueryStringValue(g, prefix+"."+goName, desc, subMsg)
	}
	if subMsg != nil {
		if subMsg.Desc.FullName() == "google.protobuf.Value" {
			return genQueryStringValue(g, prefix+"."+goName, desc, subMsg)
		}
		if desc.Kind() == protoreflect.MessageKind {
			if desc.IsList() || desc.IsMap() || desc.IsExtension() || desc.IsWeak() || desc.IsPacked() || desc.IsPlaceholder() {
				return nil
			}
		}
		name := prefix + "." + goName
		g.P("if ", name, " == nil {")
		g.P("	", name, " = &", subMsg.GoIdent, "{}")
		g.P("}")
		for _, fd := range subMsg.Fields {
			if string(fd.Desc.Name()) == names[1] {
				return genQueryString(g, name, names[1:], fd.GoName, fd.Desc, fd.Message)
			}
		}
	}
	return nil
}

func genQueryStringValue(g *protogen.GeneratedFile, path string, desc protoreflect.FieldDescriptor, subMsg *protogen.Message) error {
	switch desc.Kind() {
	case protoreflect.BoolKind:
		if desc.IsList() {
			g.P("list := make([]bool, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", strconvPackage.Ident("ParseBool"), "(text)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, val)")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", strconvPackage.Ident("ParseBool"), "(vals[0])")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			if desc.HasOptionalKeyword() {
				g.P(path, " = &val")
			} else {
				g.P(path, " = val")
			}
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if desc.IsList() {
			g.P("list := make([]int32, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", strconvPackage.Ident("ParseInt"), "(text, 10, 32)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, int32(val))")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", strconvPackage.Ident("ParseInt"), "(vals[0], 10, 32)")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			if desc.HasOptionalKeyword() {
				g.P("int32val := int32(val)")
				g.P(path, " = &int32val")
			} else {
				g.P(path, " = int32(val)")
			}
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if desc.IsList() {
			g.P("list := make([]uint32, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", strconvPackage.Ident("ParseUint"), "(text, 10, 32)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, uint32(val))")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", strconvPackage.Ident("ParseUint"), "(vals[0], 10, 32)")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			if desc.HasOptionalKeyword() {
				g.P("uint32val := uint32(val)")
				g.P(path, " = &uint32val")
			} else {
				g.P(path, " = uint32(val)")
			}
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if desc.IsList() {
			g.P("list := make([]int64, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", strconvPackage.Ident("ParseInt"), "(text, 10, 64)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, val)")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", strconvPackage.Ident("ParseInt"), "(vals[0], 10, 64)")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			if desc.HasOptionalKeyword() {
				g.P(path, " = &val")
			} else {
				g.P(path, " = val")
			}
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if desc.IsList() {
			g.P("list := make([]uint64, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", strconvPackage.Ident("ParseUint"), "(text, 10, 64)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, val)")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", strconvPackage.Ident("ParseUint"), "(vals[0], 10, 64)")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			if desc.HasOptionalKeyword() {
				g.P(path, " = &val")
			} else {
				g.P(path, " = val")
			}
		}
	case protoreflect.FloatKind:
		if desc.IsList() {
			g.P("list := make([]float32, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", strconvPackage.Ident("ParseFloat"), "(text, 32)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, float32(val))")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", strconvPackage.Ident("ParseFloat"), "(vals[0], 32)")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			if desc.HasOptionalKeyword() {
				g.P("float32val := float32(val)")
				g.P(path, " = &float32val")
			} else {
				g.P(path, " = float32(val)")
			}
		}
	case protoreflect.DoubleKind:
		if desc.IsList() {
			g.P("list := make([]float64, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", strconvPackage.Ident("ParseFloat"), "(text, 64)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, val)")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", strconvPackage.Ident("ParseFloat"), "(vals[0], 64)")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			if desc.HasOptionalKeyword() {
				g.P(path, " = &val")
			} else {
				g.P(path, " = val")
			}
		}
	case protoreflect.StringKind:
		if desc.IsList() {
			g.P(path, " = vals")
		} else {
			if desc.HasOptionalKeyword() {
				g.P(path, " = &vals[0]")
			} else {
				g.P(path, " = vals[0]")
			}
		}
	case protoreflect.BytesKind:
		if desc.IsList() {
			g.P("list := make([][]byte, 0, len(vals))")
			g.P("for _, text := range vals {")
			g.P("	val, err := ", base64Package.Ident("StdEncoding.DecodeString"), "(text)")
			g.P("	if err != nil {")
			g.P("		return err")
			g.P("	}")
			g.P("	list = append(list, val)")
			g.P("}")
			g.P(path, " = list")
		} else {
			g.P("val, err := ", base64Package.Ident("StdEncoding.DecodeString"), "(vals[0])")
			g.P("if err != nil {")
			g.P("	return err")
			g.P("}")
			g.P(path, " = val")
		}
	case protoreflect.MessageKind:
		if desc.IsList() {
			if subMsg.Desc.FullName() == "google.protobuf.Value" {
				g.P("var list []interface{}")
				g.P("for _, text := range vals {")
				g.P("	var v interface{}")
				g.P("	err := ", jsonPackage.Ident("NewDecoder"), "(", stringsPackage.Ident("NewReader"), "(text)).Decode(&v)")
				g.P("	if err != nil {")
				g.P("		list = append(list, v)")
				g.P("	} else {")
				g.P("		list = append(list, text)")
				g.P("	}")
				g.P("}")
				g.P("val, _ := ", structpbPackage.Ident("NewList(list)"))
				g.P("", path, " = ", structpbPackage.Ident("NewListValue"), "(val)")
			}
		} else {
			if subMsg.Desc.FullName() == "google.protobuf.Value" {
				g.P("if len(vals) > 1 {")
				g.P("	var list []interface{}")
				g.P("	for _, text := range vals {")
				g.P("		var v interface{}")
				g.P("		err := ", jsonPackage.Ident("NewDecoder"), "(", stringsPackage.Ident("NewReader"), "(text)).Decode(&v)")
				g.P("		if err != nil {")
				g.P("			list = append(list, v)")
				g.P("		} else {")
				g.P("			list = append(list, text)")
				g.P("		}")
				g.P("	}")
				g.P("	val, _ := ", structpbPackage.Ident("NewList(list)"))
				g.P("	", path, " = ", structpbPackage.Ident("NewListValue"), "(val)")
				g.P("} else {")
				g.P("	var v interface{}")
				g.P("	err := ", jsonPackage.Ident("NewDecoder"), "(", stringsPackage.Ident("NewReader"), "(vals[0])).Decode(&v)")
				g.P("	if err != nil {")
				g.P("		val, _ := ", structpbPackage.Ident("NewValue(v)"))
				g.P("		", path, " = val")
				g.P("	} else {")
				g.P("		", path, " = ", structpbPackage.Ident("NewStringValue"), "(vals[0])")
				g.P("	}")
				g.P("}")
			} else {
				g.P("if ", path, " == nil {")
				g.P("	", path, " = &", subMsg.GoIdent, "{}")
				g.P("}")
			}
		}
	default:
		return fmt.Errorf("not support type %q for query string", desc.Kind())
	}
	return nil
}

type queryParam struct {
	Root   *protogen.Field
	Field  *protogen.Field
	GoName string
	Name   string
}

func createQueryParams(fields []*protogen.Field) []*queryParam {
	queryParams := make([]*queryParam, 0)

	var fn func(parent *queryParam, fields []*protogen.Field, root *protogen.Field)
	fn = func(parent *queryParam, fields []*protogen.Field, root *protogen.Field) {
		if root != nil && root.Message != nil {
			if root.Message.Desc.FullName() == "google.protobuf.Value" {
				return
			}
		}
		for _, field := range fields {
			if field.Desc.Kind() == protoreflect.MessageKind {
				if field.Desc.IsList() || field.Desc.IsMap() || field.Desc.IsExtension() || field.Desc.IsWeak() || field.Desc.IsPacked() || field.Desc.IsPlaceholder() {
					continue
				}
			}
			rootField := root
			if rootField == nil {
				rootField = field
			}
			if field.Desc.Kind() == protoreflect.MessageKind {
				q := &queryParam{
					Root:   rootField,
					Field:  field,
					GoName: fmt.Sprintf("%s%s", parent.GoName, field.GoName),
					Name:   fmt.Sprintf("%s%s", parent.Name, field.Desc.Name()),
				}
				queryParams = append(queryParams, q)
				qp := *q
				qp.Name, qp.GoName = qp.Name+".", qp.GoName+"."
				fn(&qp, field.Message.Fields, rootField)
				continue
			}
			queryParams = append(queryParams, &queryParam{
				Root:   rootField,
				Field:  field,
				GoName: fmt.Sprintf("%s%s", parent.GoName, field.GoName),
				Name:   fmt.Sprintf("%s%s", parent.Name, field.Desc.Name()),
			})
		}
	}

	fn(&queryParam{GoName: "", Name: ""}, fields, nil)
	return queryParams
}
