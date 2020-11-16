package genopenapi

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/BlackCodes/micro-swagger/protoc-gen-micro-swagger/internal/descriptor"
	"github.com/BlackCodes/micro-swagger/protoc-gen-micro-swagger/options"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func GetCommentBySourceCode(s *descriptor.Service, commentType int, serviceIndex int32, serviceMethodIndex int32, messageIndex int32, messageFieldIndex int32) string {
	comment := ""

	var trimComment = func(c string) string {
		if len(c) > 0 {
			c = strings.TrimRight(c, "\n")
			c = strings.TrimSpace(c)
			c = strings.Replace(c, "\n ", "\n", -1)
		}
		//glog.V(1).Infof("return c:%v", c)
		return c
	}
	for _, item := range s.File.SourceCodeInfo.Location {
		if item.LeadingComments == nil {
			continue
		}
		path := item.GetPath()
		switch path[0] {
		case 4:
			if commentType == 4 {
				if messageFieldIndex > -1 && len(path) > 3 && path[3] == messageFieldIndex {
					comment = item.GetLeadingComments()
					return trimComment(comment)
				}
			}
		case 6:
			//glog.V(1).Infof("path:%v,commentType:%v,serviceIndex:%v,serviceMethodIndex:%v,comment:%v", path, commentType, serviceIndex, serviceMethodIndex, comment)
			if commentType == 6 {
				if serviceMethodIndex > -1 && len(path) > 3 && path[3] == serviceMethodIndex {
					comment = item.GetLeadingComments()
					return trimComment(comment)
				}
			}
		}
	}

	return comment
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func renderServices1(services []*descriptor.Service, paths openapiPathsObject, reg *descriptor.Registry, requestResponseRefs, customRefs refMap, msgs []*descriptor.Message) error {
	// Correctness of svcIdx and methIdx depends on 'services' containing the services in the same order as the 'file.Service' array.
	svcBaseIdx := 0
	var lastFile *descriptor.File = nil

	// Req和Rsp定义
	paramsMsg := make(map[string]struct{})
	for _, item := range msgs {
		if strings.HasSuffix(item.GetName(), "Rsp") || strings.HasSuffix(item.GetName(), "Req") {
			paramsMsg[item.GetName()] = struct{}{}
		}
	}
	for svcIdx, svc := range services {
		if svc.File != lastFile {
			lastFile = svc.File
			svcBaseIdx = svcIdx
		}
		tag := ""
		apiPreFix := ""
		c, _ := proto.GetExtension(svc.GetOptions(), options.E_Category)
		if _c, ok := c.(*string); ok {
			tag = *_c
		}
		preFix, _ := proto.GetExtension(svc.GetOptions(), options.E_ApiPrefix)
		if preFix, ok := preFix.(*string); ok {
			apiPreFix = *preFix
			apiPreFix = strings.Trim(apiPreFix, "/")
		}
		for methIdx, meth := range svc.Methods {
			// 定义入参
			Reqparameters := openapiParametersObject{}
			if moption, err := proto.GetExtension(meth.GetOptions(), options.E_Hkv); err == nil {
				if mop, ok := moption.(*options.OptionsMethod); ok {
					if mop.Ignore {
						continue
					}
					for k, v := range mop.HeaderMap {
						headers := openapiParameterObject{
							Name:        k,
							In:          "header",
							Required:    true,
							Type:        "string",
							Description: v,
							Format:      "string",
						}
						Reqparameters = append(Reqparameters, headers)
					}
				}

			}
			reqName := fmt.Sprintf("%sReq", meth.GetName())
			if strings.HasSuffix(meth.RequestType.GetName(), "Req") {
				reqName = meth.RequestType.GetName()
			}

			if _, ok := paramsMsg[reqName]; ok {
				paramObject := openapiParameterObject{
					Name: "Object",
					In:   "body",
				}

				for _, f := range reg.GetAllFQMNs() {
					//glog.V(1).Infof("the fqmn name:%v,methodName:%v,current reqName:%v",f,meth.GetName(),reqName)
					if strings.Contains(f, fmt.Sprintf(".%s", reqName)) {
						defname, ok := fullyQualifiedNameToOpenAPIName(f, reg)
						if ok {
							reqName = defname
						}
						break
					}
				}
				paramObject.Schema = &openapiSchemaObject{
					schemaCore: schemaCore{Ref: fmt.Sprintf("#/definitions/%s", reqName)},
				}
				paramObject.Required = true

				Reqparameters = append(Reqparameters, paramObject)
			}
			comment := GetCommentBySourceCode(svc, 6, int32(svcIdx), int32(methIdx), 0, 0)
			responseAPIs := make(openapiResponsesObject)
			rspName := fmt.Sprintf("%sRsp", meth.GetName())
			if strings.HasSuffix(meth.ResponseType.GetName(), "Rsq") {
				reqName = meth.ResponseType.GetName()
			}
			if _, ok := paramsMsg[rspName]; !ok {
				responseAPIs["200"] = openapiResponseObject{
					Description: "返回body体空",
				}
			} else {
				for _, f := range reg.GetAllFQMNs() {
					if strings.Contains(f, fmt.Sprintf(".%s", rspName)) {
						defName, ok := fullyQualifiedNameToOpenAPIName(f, reg)
						if ok {
							rspName = defName
						}
						break
					}
				}
				responseAPIs["200"] = openapiResponseObject{
					Description: "返回body体",
					Schema: openapiSchemaObject{
						schemaCore: schemaCore{Ref: fmt.Sprintf("#/definitions/%s", rspName)},
					},
				}
			}

			if len(tag) == 0 {
				tag = svc.GetName()
			}
			pathOperation := &openapiOperationObject{
				Summary:     comment,
				Description: comment,
				OperationID: fmt.Sprintf("op_%s", meth.GetName()),
				Responses:   responseAPIs,
				Parameters:  Reqparameters,
				Tags:        []string{tag},
				Produces:    []string{"application/json"},
			}
			pathObject := openapiPathItemObject{
				Post: pathOperation,
			}

			apiPath := fmt.Sprintf("/%s/%s", strings.ToLower(svc.GetName()), strings.ToLower(string(meth.GetName()[0]))+meth.GetName()[1:])
			if len(apiPreFix) > 0 {
				apiPath = fmt.Sprintf("/%s%s", apiPreFix, apiPath)
			}
			paths[apiPath] = pathObject
			glog.V(1).Infof("methodName Name:%v,req Name:%v,rsp Name %v,apiPath:%v,comment:%s", meth.GetName(), reqName, rspName, apiPath, comment)
			for bIdx, b := range meth.Bindings {
				// Iterate over all the OpenAPI parameters
				parameters := openapiParametersObject{}
				for _, parameter := range b.PathParams {

					var paramType, paramFormat, desc, collectionFormat, defaultValue string
					var enumNames []string
					var items *openapiItemsObject
					var minItems *int
					switch pt := parameter.Target.GetType(); pt {
					case descriptorpb.FieldDescriptorProto_TYPE_GROUP, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
						if descriptor.IsWellKnownType(parameter.Target.GetTypeName()) {
							if parameter.IsRepeated() {
								return fmt.Errorf("only primitive and enum types are allowed in repeated path parameters")
							}
							schema := schemaOfField(parameter.Target, reg, customRefs)
							paramType = schema.Type
							paramFormat = schema.Format
							desc = schema.Description
							defaultValue = schema.Default
						} else {
							return fmt.Errorf("only primitive and well-known types are allowed in path parameters")
						}
					case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
						enum, err := reg.LookupEnum("", parameter.Target.GetTypeName())
						if err != nil {
							return err
						}
						paramType = "string"
						paramFormat = ""
						enumNames = listEnumNames(enum)
						if reg.GetEnumsAsInts() {
							paramType = "integer"
							paramFormat = ""
							enumNames = listEnumNumbers(enum)
						}
						schema := schemaOfField(parameter.Target, reg, customRefs)
						desc = schema.Description
						defaultValue = schema.Default
					default:
						var ok bool
						paramType, paramFormat, ok = primitiveSchema(pt)
						if !ok {
							return fmt.Errorf("unknown field type %v", pt)
						}

						schema := schemaOfField(parameter.Target, reg, customRefs)
						desc = schema.Description
						defaultValue = schema.Default
					}

					if parameter.IsRepeated() {
						core := schemaCore{Type: paramType, Format: paramFormat}
						if parameter.IsEnum() {
							var s []string
							core.Enum = enumNames
							enumNames = s
						}
						items = (*openapiItemsObject)(&core)
						paramType = "array"
						paramFormat = ""
						collectionFormat = reg.GetRepeatedPathParamSeparatorName()
						minItems = new(int)
						*minItems = 1
					}

					if desc == "" {
						desc = fieldProtoComments(reg, parameter.Target.Message, parameter.Target)
					}
					parameterString := parameter.String()
					if reg.GetUseJSONNamesForFields() {
						parameterString = lowerCamelCase(parameterString, meth.RequestType.Fields, msgs)
					}
					parameters = append(parameters, openapiParameterObject{
						Name:        parameterString,
						Description: desc,
						In:          "path",
						Required:    true,
						Default:     defaultValue,
						// Parameters in gRPC-Gateway can only be strings?
						Type:             paramType,
						Format:           paramFormat,
						Enum:             enumNames,
						Items:            items,
						CollectionFormat: collectionFormat,
						MinItems:         minItems,
					})
				}
				// Now check if there is a body parameter
				if b.Body != nil {
					var schema openapiSchemaObject
					desc := ""

					if len(b.Body.FieldPath) == 0 {
						schema = openapiSchemaObject{
							schemaCore: schemaCore{},
						}

						wknSchemaCore, isWkn := wktSchemas[meth.RequestType.FQMN()]
						if !isWkn {
							err := schema.setRefFromFQN(meth.RequestType.FQMN(), reg)
							if err != nil {
								return err
							}
						} else {
							schema.schemaCore = wknSchemaCore

							// Special workaround for Empty: it's well-known type but wknSchemas only returns schema.schemaCore; but we need to set schema.Properties which is a level higher.
							if meth.RequestType.FQMN() == ".google.protobuf.Empty" {
								schema.Properties = &openapiSchemaObjectProperties{}
							}
						}
					} else {
						lastField := b.Body.FieldPath[len(b.Body.FieldPath)-1]
						schema = schemaOfField(lastField.Target, reg, customRefs)
						if schema.Description != "" {
							desc = schema.Description
						} else {
							desc = fieldProtoComments(reg, lastField.Target.Message, lastField.Target)
						}
					}

					if meth.GetClientStreaming() {
						desc += " (streaming inputs)"
					}
					parameters = append(parameters, openapiParameterObject{
						Name:        "body",
						Description: desc,
						In:          "body",
						Required:    true,
						Schema:      &schema,
					})
					// add the parameters to the query string
					queryParams, err := messageToQueryParameters(meth.RequestType, reg, b.PathParams, b.Body)
					if err != nil {
						return err
					}
					parameters = append(parameters, queryParams...)
				} else if b.HTTPMethod == "GET" || b.HTTPMethod == "DELETE" {
					// add the parameters to the query string
					queryParams, err := messageToQueryParameters(meth.RequestType, reg, b.PathParams, b.Body)
					if err != nil {
						return err
					}
					parameters = append(parameters, queryParams...)
				}

				pathItemObject, ok := paths[templateToOpenAPIPath(b.PathTmpl.Template, reg, meth.RequestType.Fields, msgs)]
				if !ok {
					pathItemObject = openapiPathItemObject{}
				}

				methProtoPath := protoPathIndex(reflect.TypeOf((*descriptorpb.ServiceDescriptorProto)(nil)), "Method")
				desc := "A successful response."
				var responseSchema openapiSchemaObject

				if b.ResponseBody == nil || len(b.ResponseBody.FieldPath) == 0 {
					responseSchema = openapiSchemaObject{
						schemaCore: schemaCore{},
					}

					// Don't link to a full definition for
					// empty; it's overly verbose.
					// schema.Properties{} renders it as
					// well, without a definition
					wknSchemaCore, isWkn := wktSchemas[meth.ResponseType.FQMN()]
					if !isWkn {
						err := responseSchema.setRefFromFQN(meth.ResponseType.FQMN(), reg)
						if err != nil {
							return err
						}
					} else {
						responseSchema.schemaCore = wknSchemaCore

						// Special workaround for Empty: it's well-known type but wknSchemas only returns schema.schemaCore; but we need to set schema.Properties which is a level higher.
						if meth.ResponseType.FQMN() == ".google.protobuf.Empty" {
							responseSchema.Properties = &openapiSchemaObjectProperties{}
						}
					}
				} else {
					// This is resolving the value of response_body in the google.api.HttpRule
					lastField := b.ResponseBody.FieldPath[len(b.ResponseBody.FieldPath)-1]
					responseSchema = schemaOfField(lastField.Target, reg, customRefs)
					if responseSchema.Description != "" {
						desc = responseSchema.Description
					} else {
						desc = fieldProtoComments(reg, lastField.Target.Message, lastField.Target)
					}
				}
				if meth.GetServerStreaming() {
					desc += "(streaming responses)"
					responseSchema.Type = "object"
					swgRef, _ := fullyQualifiedNameToOpenAPIName(meth.ResponseType.FQMN(), reg)
					responseSchema.Title = "Stream result of " + swgRef

					props := openapiSchemaObjectProperties{
						keyVal{
							Key: "result",
							Value: openapiSchemaObject{
								schemaCore: schemaCore{
									Ref: responseSchema.Ref,
								},
							},
						},
					}
					statusDef, hasStatus := fullyQualifiedNameToOpenAPIName(".google.rpc.Status", reg)
					if hasStatus {
						props = append(props, keyVal{
							Key: "error",
							Value: openapiSchemaObject{
								schemaCore: schemaCore{
									Ref: fmt.Sprintf("#/definitions/%s", statusDef)},
							},
						})
					}
					responseSchema.Properties = &props
					responseSchema.Ref = ""
				}

				tag := svc.GetName()
				if pkg := svc.File.GetPackage(); pkg != "" && reg.IsIncludePackageInTags() {
					tag = pkg + "." + tag
				}

				operationObject := &openapiOperationObject{
					Tags:       []string{tag},
					Parameters: parameters,
					Responses: openapiResponsesObject{
						"200": openapiResponseObject{
							Description: desc,
							Schema:      responseSchema,
						},
					},
				}
				if !reg.GetDisableDefaultErrors() {
					errDef, hasErrDef := fullyQualifiedNameToOpenAPIName(".google.rpc.Status", reg)
					if hasErrDef {
						// https://github.com/OAI/OpenAPI-Specification/blob/3.0.0/versions/2.0.md#responses-object
						operationObject.Responses["default"] = openapiResponseObject{
							Description: "An unexpected error response.",
							Schema: openapiSchemaObject{
								schemaCore: schemaCore{
									Ref: fmt.Sprintf("#/definitions/%s", errDef),
								},
							},
						}
					}
				}
				operationObject.OperationID = fmt.Sprintf("%s_%s", svc.GetName(), meth.GetName())
				if reg.GetSimpleOperationIDs() {
					operationObject.OperationID = meth.GetName()
				}
				if bIdx != 0 {
					// OperationID must be unique in an OpenAPI v2 definition.
					operationObject.OperationID += strconv.Itoa(bIdx + 1)
				}

				// Fill reference map with referenced request messages
				for _, param := range operationObject.Parameters {
					if param.Schema != nil && param.Schema.Ref != "" {
						requestResponseRefs[param.Schema.Ref] = struct{}{}
					}
				}

				methComments := protoComments(reg, svc.File, nil, "Service", int32(svcIdx-svcBaseIdx), methProtoPath, int32(methIdx))
				if err := updateOpenAPIDataFromComments(reg, operationObject, meth, methComments, false); err != nil {
					panic(err)
				}

				opts, err := getMethodOpenAPIOption(reg, meth)
				if opts != nil {
					if err != nil {
						panic(err)
					}
					operationObject.Externaloptionss = protoExternalDocumentationToOpenAPIExternalDocumentation(opts.ExternalOptionss, reg, meth)
					// TODO(ivucica): this would be better supported by looking whether the method is deprecated in the proto file
					operationObject.Deprecated = opts.Deprecated

					if opts.Summary != "" {
						operationObject.Summary = opts.Summary
					}
					if opts.Description != "" {
						operationObject.Description = opts.Description
					}
					if len(opts.Tags) > 0 {
						operationObject.Tags = make([]string, len(opts.Tags))
						copy(operationObject.Tags, opts.Tags)
					}
					if opts.OperationId != "" {
						operationObject.OperationID = opts.OperationId
					}
					if opts.Security != nil {
						newSecurity := []openapiSecurityRequirementObject{}
						if operationObject.Security != nil {
							newSecurity = *operationObject.Security
						}
						for _, secReq := range opts.Security {
							newSecReq := openapiSecurityRequirementObject{}
							for secReqKey, secReqValue := range secReq.SecurityRequirement {
								if secReqValue == nil {
									continue
								}

								newSecReqValue := make([]string, len(secReqValue.Scope))
								copy(newSecReqValue, secReqValue.Scope)
								newSecReq[secReqKey] = newSecReqValue
							}

							if len(newSecReq) > 0 {
								newSecurity = append(newSecurity, newSecReq)
							}
						}
						operationObject.Security = &newSecurity
					}
					if opts.Responses != nil {
						for name, resp := range opts.Responses {
							// Merge response data into default response if available.
							respObj := operationObject.Responses[name]
							if resp.Description != "" {
								respObj.Description = resp.Description
							}
							if resp.Schema != nil {
								respObj.Schema = openapiSchemaFromProtoSchema(resp.Schema, reg, customRefs, meth)
							}
							if resp.Examples != nil {
								respObj.Examples = openapiExamplesFromProtoExamples(resp.Examples)
							}
							if resp.Extensions != nil {
								exts, err := processExtensions(resp.Extensions)
								if err != nil {
									return err
								}
								respObj.extensions = exts
							}
							operationObject.Responses[name] = respObj
						}
					}

					if opts.Extensions != nil {
						exts, err := processExtensions(opts.Extensions)
						if err != nil {
							return err
						}
						operationObject.extensions = exts
					}

					if len(opts.Produces) > 0 {
						operationObject.Produces = make([]string, len(opts.Produces))
						copy(operationObject.Produces, opts.Produces)
					}

					// TODO(ivucica): add remaining fields of operation object
				}

				switch b.HTTPMethod {
				case "DELETE":
					pathItemObject.Delete = operationObject
				case "GET":
					pathItemObject.Get = operationObject
				case "POST":
					pathItemObject.Post = operationObject
				case "PUT":
					pathItemObject.Put = operationObject
				case "PATCH":
					pathItemObject.Patch = operationObject
				}
				paths[templateToOpenAPIPath(b.PathTmpl.Template, reg, meth.RequestType.Fields, msgs)] = pathItemObject
			}
		}
	}

	// Success! return nil on the error object
	return nil
}
