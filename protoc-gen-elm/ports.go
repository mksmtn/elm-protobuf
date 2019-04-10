package main

import (
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func (elmFg *FileGenerator) GenerateElmPortsForMethod(prefix string, serviceName string, methodName string, rawInputType string) error {
	serviceNameLowerCase := strings.ToLower(serviceName)
	portsNameCommonPart := prefix + serviceName + methodName
	outPortName := "sub" + portsNameCommonPart
	inPortName := "cmd" + portsNameCommonPart
	inputType := strings.ReplaceAll(rawInputType, "." + serviceNameLowerCase + ".", "") 

	elmFg.P("")
	elmFg.P("")
	elmFg.P("port %s : (JE.Value -> msg) -> Sub msg", outPortName)
	elmFg.P("")
	elmFg.P("port %s : RequestOptions %s -> Cmd msg", inPortName, inputType)

	return nil
}

func elmMsg(methodIndex int, methodName string) string {
	msgName := "    "

	if methodIndex == 0 {
		msgName += "="
	} else {
		msgName += "|"
	}

	msgName += " " + methodName + " JE.Value"

	return msgName
}

func (elmFg *FileGenerator) GenerateElmPorts(prefix string, inFile *descriptor.FileDescriptorProto) error {

	for _, inService := range inFile.GetService() {

		serviceName := inService.GetName()

		elmFg.P("")
		elmFg.P("")
		elmFg.P("type alias RequestOptions request = { url : String, request : request }")
		elmFg.P("")
		elmFg.P("")
		elmFg.P("type %s", serviceName + "Msg")

		// msg
		for methodIndex, inMethod := range inService.GetMethod() {
			msg := elmMsg(methodIndex, inMethod.GetName())
			elmFg.P(msg)
		}

		// ports
		for _, inMethod := range inService.GetMethod() {
			elmFg.GenerateElmPortsForMethod(prefix, serviceName, inMethod.GetName(), inMethod.GetInputType())
		}
	}

	return nil
}

func (jsFg *FileGenerator) GenerateJsPorts(prefix string, inFile *descriptor.FileDescriptorProto) error {
	for _, inService := range inFile.GetService() {

		serviceName := inService.GetName()
		serviceNameLowerCase := strings.ToLower(serviceName)

		jsFg.P("const { %sClient } = require('./%s_grpc_web_pb');", serviceName, serviceNameLowerCase)
		jsFg.P("const {")

		for _, inMessage := range inFile.GetMessageType() {
			jsFg.P("  %s,", inMessage.GetName())
		}

		jsFg.P("} = require('./%s_pb');", serviceNameLowerCase)

	}

	jsFg.P("const init = (app) => {")

	for _, inService := range inFile.GetService() {
		serviceName := inService.GetName()
		serviceNameLowerCase := strings.ToLower(serviceName)
		for _, inMethod := range inService.GetMethod() {
			methodName := inMethod.GetName()
			methodNameLowerCase := firstLower(methodName)
			inputType := strings.ReplaceAll(inMethod.GetInputType(), "." + serviceNameLowerCase + ".", "") 
			jsFg.P("  app.ports.cmd%s%s.subscribe((options) => {", serviceName, methodName)
			jsFg.P("    const service = new %sClient(options.url);", serviceName)
			jsFg.P("    const request = new %s(options.request);", inputType)
			jsFg.P("    service.%s(request, {}, (err, response) => {", methodNameLowerCase)
			jsFg.P("      const result = {")
			jsFg.P("        err,")
			jsFg.P("        response,")
			jsFg.P("        type: '%s.%s'", serviceName, methodNameLowerCase)
			jsFg.P("      };")
			jsFg.P("      app.ports.sub%s%s.send(result);", serviceName, methodName)
			jsFg.P("    });")
			jsFg.P("  });")
		}
	}

	jsFg.P("};")
	jsFg.P("")
	jsFg.P("export const ElmGRPC = {")
	jsFg.P("  init")
	jsFg.P("};")
	jsFg.P("")
	jsFg.P("window.ElmGRPC = ElmGRPC;")

	return nil
}
