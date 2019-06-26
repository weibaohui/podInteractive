#
本项目演示了如何通过web方式获取pod运行日志
以及进入pod执行命令

#quickstart
1. kubectl apply -f deploy/deploy.yaml
2. 访问http://nodeIP:nodePort
3. 输入namespace、podName、containerName
4. 点击生成日志连接、生成exec连接
5. 访问

#集成
可以参考示例，将websocket接入到项目中

#截图
![Log日志](https://github.com/weibaohui/podInteractive/blob/master/images/log.png)

