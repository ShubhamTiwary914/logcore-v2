FROM alpine:3.22

WORKDIR /app

COPY pub.js .

#k6 with xk6-mqtt support
RUN apk add --no-cache curl \
    && curl -L https://github.com/ShubhamTiwary914/dotfiles/raw/main/bin/k6 -o ./k6 \
    && chmod +x ./k6

#default envs
ENV RELATION=weather_sensor
ENV INTERVAL=1
ENV DURATION=10
ENV VUS=10

#gotta pass "VERNE_IP" env during container creation -> points to the K8s load balancer for verneMQTT
ENTRYPOINT [ "sh", "-c", "RELATION=$RELATION DURATION=$DURATION INTERVAL=$INTERVAL VERNE_IP=$VERNE_IP VUS=$VUS /app/k6 run --vus $VUS --duration ${DURATION}s /app/pub.js" ]
