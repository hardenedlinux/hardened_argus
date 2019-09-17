#服务端  
配置，通过配置文件（json格式）配置  
```
{
	"logdir":"/Users/root/tmp/argus/log",
	"dbdir":"/Users/root/tmp/argus/db",
	"address":"127.0.0.1:7890"
}
logdir :日志目录，脚本执行的结果返回到这个目录，文件名格式为：uid_guid_taskid_timestamp
dbdir  :数据库目录
address:接口，目录
```   
#配置接口  
1.添加任务
http://127.0.0.1:7890/manger/addtask.do?url=&guids=guid1|guid2|guid3&period=0
参数  ：
url 脚本下载地址，同时默认表示脚本的签名文件下载地址为 url.sig,即，脚本下载地址+.sig，必选参数  
guids 要执行脚本的guid列表 必选参数  
period 执行周期，单位为分钟，即从发布开始，经过period分钟之后会再次执行，0表示只执行一次，可选参数  
例如：  
http://127.0.0.1:7890/manger/addtask.do?url=http://localhost/demo.sh&guids=833831ecb5360a12c7a7b9efc0551182  
```
{"taskid":"3996a187d62daa9420f4877bfda514a5","taskurl":"http://localhost/demo.sh","period":0}
```
2.查询任务  
http://127.0.0.1:7890/manger/enumtask.do 无参数  
返回值 tasks  包含所有任务信息的数组
	  policy 任务guid数组，和tasks数组一一对应
```
{"tasks":[{"taskid":"3996a187d62daa9420f4877bfda514a5","taskurl":"http://localhost/demo.sh","period":0}],"policy":[["833831ecb5360a12c7a7b9efc0551182"]]}
```
3.查询机器信息  
http://127.0.0.1:7890/manger/enummachine.do
```
[{"uid":"clivebi","guid":"833831ecb5360a12c7a7b9efc0551182","label":"","last_active":1567999241}]
```
4.删除任务  
参数  
taskid 任务id，add返回的任务信息中包含任务ID
http://127.0.0.1:7890/manger/deltask.do?taskid=ff6dcdc235564b5adab6dae46e863f50  
#客户端  
通过环境变量配置  
ARGUS_UID 用户标志  必须配置  
ARGUS_LABEL 机器标签 可选  
ARGUS_SCRIPT_ROOT 下载和执行脚本的目录 可选，没有配置的情况下，使用home/argus_script 目录  

#签名工具
请执行argustool 查看帮助  
