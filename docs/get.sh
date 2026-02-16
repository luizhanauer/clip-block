#!/bin/sh
set -e

# Configurações do seu repositório
REPO="luizhanauer/clip-block"
BINARY_NAME="clip-block"

ARCH=$(uname -m)
case $ARCH in
    x86_64) ASSET_ARCH="amd64" ;;
    aarch64) ASSET_ARCH="arm64" ;;
    *) echo "❌ Arquitetura não suportada: $ARCH"; exit 1 ;;
esac

echo ">>> 📦 Instalador via Rede: ClipBlock"
echo ">>> Fonte: https://github.com/$REPO"

# Verifica curl e tar
if ! command -v curl >/dev/null; then echo "❌ Erro: 'curl' necessário."; exit 1; fi
if ! command -v tar >/dev/null; then echo "❌ Erro: 'tar' necessário."; exit 1; fi

TMP_DIR=$(mktemp -d)
FILENAME="${BINARY_NAME}_linux_${ASSET_ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/latest/download/${FILENAME}"

echo ">>> ⬇️  Baixando release..."
if ! curl -f -L "$URL" -o "$TMP_DIR/$FILENAME"; then
    echo "❌ Erro ao baixar release. Verifique se a tag de release existe no GitHub."
    rm -rf "$TMP_DIR"
    exit 1
fi

echo ">>> 📂 Extraindo..."
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

echo ">>> 🚀 Iniciando script de instalação..."
cd "$TMP_DIR"
chmod +x install.sh
sh ./install.sh

# Limpeza
cd - > /dev/null
rm -rf "$TMP_DIR"
echo ">>> ✅ Setup finalizado."