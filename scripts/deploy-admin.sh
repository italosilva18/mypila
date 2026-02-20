#!/bin/bash

# Script de deploy do MyPila Admin
# Este script faz o build do admin e atualiza o docker-compose

echo "🚀 Iniciando deploy do MyPila Admin..."

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Verificar se está no diretório correto
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}❌ Erro: docker-compose.yml não encontrado${NC}"
    echo "Execute este script do diretório /root/apps/mypila"
    exit 1
fi

# Verificar se o diretório admin existe
if [ ! -d "admin" ]; then
    echo -e "${RED}❌ Erro: Diretório admin não encontrado${NC}"
    exit 1
fi

echo -e "${YELLOW}📦 Fazendo build da imagem do admin...${NC}"
cd admin

# Build da imagem
docker build -t mypila-admin:latest . 2>&1

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Erro no build da imagem${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Imagem buildada com sucesso!${NC}"

cd ..

# Verificar se o serviço admin já existe no docker-compose
if ! grep -q "mypila-admin" docker-compose.yml; then
    echo -e "${YELLOW}📝 Adicionando serviço admin ao docker-compose.yml...${NC}"
    
    cat >> docker-compose.yml << 'EOF'

  # Admin Panel
  admin:
    image: mypila-admin:latest
    container_name: mypila-admin
    restart: unless-stopped
    ports:
      - "8008:80"
    depends_on:
      backend:
        condition: service_healthy
    networks:
      - m2m-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:80/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '0.25'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s
EOF
    echo -e "${GREEN}✅ Serviço adicionado ao docker-compose.yml${NC}"
else
    echo -e "${YELLOW}ℹ️ Serviço admin já existe no docker-compose.yml${NC}"
fi

# Subir o container
echo -e "${YELLOW}🚀 Subindo container do admin...${NC}"
docker-compose up -d admin 2>&1

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Erro ao subir o container${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Container do admin iniciado!${NC}"

# Verificar status
echo -e "${YELLOW}⏳ Aguardando healthcheck...${NC}"
sleep 10

HEALTH=$(docker inspect --format='{{.State.Health.Status}}' mypila-admin 2>/dev/null)
if [ "$HEALTH" = "healthy" ]; then
    echo -e "${GREEN}✅ Admin está healthy!${NC}"
elif [ "$HEALTH" = "starting" ]; then
    echo -e "${YELLOW}⏳ Admin ainda iniciando...${NC}"
else
    echo -e "${YELLOW}⚠️ Status do healthcheck: $HEALTH${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Deploy do MyPila Admin concluído!${NC}"
echo ""
echo "📊 Acesse:"
echo "   Local: http://localhost:8008"
echo "   (Adicione ao nginx para HTTPS)"
echo ""
echo "🔧 Comandos úteis:"
echo "   Logs: docker logs -f mypila-admin"
echo "   Stop: docker-compose stop admin"
echo "   Restart: docker-compose restart admin"
