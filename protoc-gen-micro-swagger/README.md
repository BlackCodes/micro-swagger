## 如何用`micro-swagger`生成在线文档
 proto-gen-micro-swagger将proto原文件生成标准的swagger文件，当前可以为go-micro版本完全匹配，支持micro的api和srv接口定义，一要需要注意，接口入参和返回参数一定是req(request)和rsp(response)

### 准备
- cd protoc-gen-micro-swagger && go build
- 将文件 `protoc-gen-micro-swagger`放入到/usr/local/bin或者系统path中；
- 执行 chmod +x protoc-gen-micro-swagger

- proto扩展文件`microOption.proto`,需要将这个文件放入到**$GOPATH/github.com/BlackCodes/micro-swagger/protoc-gen-micro-swagger/options/** 下面
- swagger服务器，默认地址：127.0.0.1，端口：9099,服务器主要用来保存swagger生成的代码，为yapi服务提供调用，如只需本地生成，此后以几项ignore.
- 注册文档服务器帐号,文档服务器使用Y-Api，去哪儿开源的一个牛xx文档管理工具 https://github.com/ymfe/yapi

### 开发

1. 将proto扩展文件引入到pb文件中；

   ```protobuf
   syntax = "proto3";
   import "github.com/BlackCodes/micro-swagger/protoc-gen-micro-swagger/options/microOption.proto";
   ```

2. 定义service中的选项

   ```protobuf
   service User {
       option (options.category) = "接口所属分类";
       option (options.apiPrefix) = "接口前缀";
   }
   ```

3. 定义具体接口的方法

   ```protobuf
   service User{
    		// 此处为接口名称（必填）
       rpc GetCaptcha(go.api.Request) returns(go.api.Response) {
       	option (options.hkv) = {
		 // 设置此接口header中的必字段
		headerMap:[{key:"字段名",value:"默认值"}]
         };
       }
   }
   ```

4. 定义接口的入参。特别注意：入参名称一定是 **方法名+Req** ； 字段注释统一在字段上方；

   > optionsRequired是否必填，optionsDefault：字段默认值

   ```protobuf
   message GetCaptchaReq{
   	// 字段注释（必填）
   	string phone = 1[(options.field)={optionsRequired:true, optionsDefault:"0"}];
   }
   ```

5. 定义接口出参。特别注意，出参名一定是 **方法名+Rsp**； 字段注释统一在字段上方；

   ```protobuf
   message GetCaptchaRsp {
   	// 字段注释（必填）
   	int64 Captcha = 1;
   }
   ```

### shell编译pb文件

> shell中编译pb文件
>
> 参数：--micro-swagger_out=doc_server=127.0.0.1,doc_server_port=9099,project=XXproject:.
>
> doc_server:文档服务器地址(提供swagger服务的端口)
>
> doc_server_port：文档服务器端口(提供swagger服务的端口)
>
> project：项目名称

如果无需将编译结果送给文档服务器，以上三个参数都可不填；

```shell
protoc --proto_path=${GOPATH}/src --proto_path=.  --micro-swagger_out=doc_server=127.0.0.1,doc_server_port=9099,project=XXproject:. proto/*.proto

```

### 完整pb示例

```protobuf
syntax = "proto3";
import "github.com/BlackCodes/micro-swagger/protoc-gen-micro-swagger/options/microOption.proto";

service User {
    option (options.category) = "接口所属分类";
    option (options.apiPrefix) = "接口前缀";

    // 此处为接口名称（必填）
    rpc GetCaptcha(go.api.Request) returns(go.api.Response) {
    	option (options.hkv) = {
              // 设置此接口header中的必字段
              headerMap:[{key:"字段名",value:"默认值"}]
         };
    }
}

message GetCaptchaReq{
	// 字段注释（必填）
	string phone = 1[(options.field)={optionsRequired:true, optionsDefault:"0"}];
}

message GetCaptchaRsp {
	// 字段注释（必填）
	int64 Captcha = 1;
}
```




