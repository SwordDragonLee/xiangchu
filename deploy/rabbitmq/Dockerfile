# RabbitMQ + 延迟消息插件（订单超时依赖 x-delayed-message 交换机）
# 插件文件需预先下载到本目录（官方镜像内无 wget/curl，无法在构建时联网下载）
FROM rabbitmq:3.13-management
COPY rabbitmq_delayed_message_exchange-3.13.0.ez /opt/rabbitmq/plugins/
RUN chmod 644 /opt/rabbitmq/plugins/rabbitmq_delayed_message_exchange-3.13.0.ez && \
    rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange
