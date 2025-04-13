FROM gradle:8.13.0-jdk21 AS builder
WORKDIR /app
COPY . .
RUN gradle build --no-daemon

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=builder /app/build/libs/*.jar server.jar
EXPOSE 25565
CMD ["java", "-jar", "server.jar"]
