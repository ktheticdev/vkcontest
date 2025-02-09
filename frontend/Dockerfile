# Используем Node для сборки приложения
FROM node:18-alpine AS builder
WORKDIR /app
COPY package.json tsconfig.json vite.config.ts ./
COPY index.html ./
COPY src ./src
RUN npm install
RUN npm run build

# Используем nginx для сервировки статики
FROM nginx:stable-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
