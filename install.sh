#!/bin/sh
set -e

# Garante que o script rode dentro da pasta extraída
cd "$(dirname "$0")"

# Caminhos Padrão do Linux
BIN_DIR="$HOME/.local/bin"
AUTOSTART_DIR="$HOME/.config/autostart"
APPS_DIR="$HOME/.local/share/applications"
ICONS_DIR="$HOME/.local/share/icons"

FINAL_BIN="$BIN_DIR/clip-block"

echo ">>> 🔒 Verificando dependências..."
if ! command -v xclip >/dev/null 2>&1; then
    echo ">>> 📦 Instalando xclip (necessário para clipboard)..."
    if [ "$(id -u)" -eq 0 ]; then
        apt-get update && apt-get install -y xclip
    else
        sudo apt-get update && sudo apt-get install -y xclip
    fi
fi

echo ">>> 🚀 Instalando ClipBlock..."

# 1. Cria diretórios necessários
mkdir -p "$BIN_DIR"
mkdir -p "$AUTOSTART_DIR"
mkdir -p "$APPS_DIR"
mkdir -p "$ICONS_DIR"

# 2. Copia o binário
if [ -f "bin/clip-block" ]; then
    cp bin/clip-block "$FINAL_BIN"
    chmod +x "$FINAL_BIN"
else
    echo "❌ Erro: Binário não encontrado."
    exit 1
fi

# 3. Instala o Ícone (Se existir no pacote)
ICON_NAME="accessor-clipboard" # Ícone padrão do sistema (fallback)
if [ -f "appicon.png" ]; then
    cp appicon.png "$ICONS_DIR/clip-block.png"
    ICON_NAME="$ICONS_DIR/clip-block.png"
fi

# 4. Configura Atalho F9
NAME="ClipBlock Toggle"
BINDING="F9"
CUSTOM_PATH="/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/custom0/"
CURRENT_LIST=$(gsettings get org.gnome.settings-daemon.plugins.media-keys custom-keybindings)

if [ "$CURRENT_LIST" = "@as []" ] || [ "$CURRENT_LIST" = "[]" ]; then
    NEW_LIST="['$CUSTOM_PATH']"
else
    # Adiciona apenas se não existir
    if echo "$CURRENT_LIST" | grep -q "$CUSTOM_PATH"; then
        NEW_LIST="$CURRENT_LIST"
    else
        NEW_LIST=$(echo "$CURRENT_LIST" | sed "s/\]$/ , '$CUSTOM_PATH'\]/")
    fi
fi

# Aplica gsettings (ignora erro se schema não existir em distros não-gnome)
gsettings set org.gnome.settings-daemon.plugins.media-keys custom-keybindings "$NEW_LIST" 2>/dev/null || true
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$CUSTOM_PATH name "$NAME" 2>/dev/null || true
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$CUSTOM_PATH command "$FINAL_BIN" 2>/dev/null || true
gsettings set org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$CUSTOM_PATH binding "$BINDING" 2>/dev/null || true

# 5. Criação do Arquivo .desktop (Menu + Autostart)
# Vamos usar o mesmo arquivo para ambos os locais para garantir consistência

cat <<EOF > /tmp/clipblock.desktop
[Desktop Entry]
Type=Application
Name=ClipBlock
Comment=Gerenciador de Clipboard Invisível
Exec=$FINAL_BIN
Icon=$ICON_NAME
Terminal=false
Categories=Utility;Development;
Keywords=clipboard;manager;copy;paste;
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
EOF

# Instala no Menu de Aplicativos
cp /tmp/clipblock.desktop "$APPS_DIR/clipblock.desktop"

# Instala no Autostart (Inicialização)
mv /tmp/clipblock.desktop "$AUTOSTART_DIR/clipblock.desktop"

# 6. Atualiza banco de dados de desktop (para aparecer na hora)
update-desktop-database "$APPS_DIR" 2>/dev/null || true

echo "--------------------------------------------------------"
echo "✅ ClipBlock instalado com sucesso!"
echo "📂 Menu: Disponível em 'Mostrar Aplicativos'"
echo "⌨️  Atalho: F9 para alternar"
echo "--------------------------------------------------------"