#!/bin/sh
set -e

# Garante que o script rode dentro da pasta extraída
cd "$(dirname "$0")"

BIN_DIR="$HOME/.local/bin"
AUTOSTART_DIR="$HOME/.config/autostart"
FINAL_BIN="$BIN_DIR/clip-block"

echo ">>> 🔒 Verificando dependências de sistema..."

# Verifica se o xclip existe, se não, instala
if ! command -v xclip >/dev/null 2>&1; then
    echo ">>> 📦 Dependência 'xclip' não encontrada. Instalando..."
    if [ "$(id -u)" -eq 0 ]; then
        apt-get update && apt-get install -y xclip
    else
        sudo apt-get update && sudo apt-get install -y xclip
    fi
else
    echo "✅ xclip já está instalado."
fi

echo ">>> 🚀 Instalando ClipBlock em $BIN_DIR..."

mkdir -p "$BIN_DIR"
mkdir -p "$AUTOSTART_DIR"

# 1. Copia o binário
if [ -f "bin/clip-block" ]; then
    cp bin/clip-block "$FINAL_BIN"
    chmod +x "$FINAL_BIN"
else
    echo "❌ Erro: Binário bin/clip-block não encontrado."
    exit 1
fi

# 2. Configura Atalho F9 (gsettings)
echo ">>> ⌨️ Configurando atalho F9..."
NAME="ClipBlock Toggle"
BINDING="F9"
CUSTOM_PATH="/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/custom0/"

CURRENT_LIST=$(gsettings get org.gnome.settings-daemon.plugins.media-keys custom-keybindings)
if [ "$CURRENT_LIST" = "@as []" ] || [ "$CURRENT_LIST" = "[]" ]; then
    NEW_LIST="['$CUSTOM_PATH']"
else
    # Lógica simples para adicionar à lista sem quebrar o formato gsettings
    NEW_LIST=$(echo "$CURRENT_LIST" | sed "s/\]$/ , '$CUSTOM_PATH'\]/")
fi

gsettings set org.gnome.settings-daemon.plugins.media-keys custom-keybindings "$NEW_LIST" || true
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$CUSTOM_PATH name "$NAME"
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$CUSTOM_PATH command "$FINAL_BIN"
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$CUSTOM_PATH binding "$BINDING"

# 3. Configura Inicialização Automática
echo ">>> 🔄 Configurando inicialização automática..."
cat <<EOF > "$AUTOSTART_DIR/clipblock.desktop"
[Desktop Entry]
Type=Application
Exec=$FINAL_BIN
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
Name=ClipBlock
Comment=Gerenciador de Clipboard
Icon=utility-terminal
EOF
chmod +x "$AUTOSTART_DIR/clipblock.desktop"

echo "--------------------------------------------------------"
echo "✅ ClipBlock instalado com sucesso!"
echo "⌨️  Aperte F9 para mostrar/esconder."
echo "--------------------------------------------------------"