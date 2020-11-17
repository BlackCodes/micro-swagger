### micro-swagger
 项目的目的很单纯，只想帮开发者解决开发接口时，还要苦逼的写文档😱，我实在折腾够了！**有个时候写的文档，不是掉了个s，就是字段类型搞错了，要么就是接口更新了，文档没来得及更新**，天天被各路神仙叫😡 😡 ，所以有病得治了，出于各种目的，就折腾了这么一个<span style="color:red">**福利**</span>，分享给曾经有这些烦劳的Coder们，喜欢的点点★★★★★；不喜欢的继续拍砖🤡
> 包含了两个项目👀
 - protoc-gen-micro-swagger：是用来配合本地pb生成swagger文件，并将文件上传到swagger服务器上；
 - web:是一个web服务器，用来提供存储swagger数据，和提供swagger查询的服务；
 - 每一次pb文件更新后，重新编绎后，都会同步到web服务；
 - 无侵入式任何项目代码

 具体用法，查看readme,有bug,烦请issue啦~🙏-
#### 最终效果图：
![image](https://github.com/BlackCodes/go-micro-swagger/blob/master/images/PREVIEW.png)
#### 运行截图
![image](https://github.com/BlackCodes/go-micro-swagger/blob/master/images/LUNCH.png)
