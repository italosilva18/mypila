# Nginx Alpine - container mínimo para landing page
FROM nginx:alpine

# Copiar arquivos estáticos
COPY dist/ /usr/share/nginx/html/

# Configuração nginx otimizada
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80