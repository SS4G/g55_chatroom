# 启动redis 进程 并且将host的12000端口映射到 容器的6379上
# 本地测试命令 redis-cli -h 127.0.0.1 -p 12000
docker run --name g55_chatroom_redis -d -p 12000:6379 redis

# redis cli 测试