#
本项目利用k8s接口实现了web shell
可以查看日志
可以进入容器执行命令

也可以通过配置ENV DOCKER_API_ADDRESS=ip:port 访问docker container
不连接docker 可以不配置

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
![Exec](https://github.com/weibaohui/podInteractive/blob/master/images/exec.png)

