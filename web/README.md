### swagger-web的角色是什么？
 swagger-web提供了两个功能：
 - 接收protoc-gen-micro生成的文件，并合并保存；
 - 对y-api提供一个查询swagger文档的接口；http://domain:port/push/get/:project,其中`:project`就是生成pb文件时指定的project参数值；

### 怎么用？
 - git clone https://github.com/BlackCodes/go-micro-swagger.git
 - go run main.go

### 参数解释？
   basePath:指定上传上来的swagger数据存放的位置；