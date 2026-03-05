# --- Variables de Configuración ---
BINARY_NAME=sms-viewer
SOURCE_FILE=main.go
BUILD_DIR=bin
DIST_DIR=dist

# Flags para optimizar el binario (reducir tamaño quitando info de debug)
LDFLAGS=-ldflags "-s -w"

# Define el objetivo por defecto al escribir 'make'
.DEFAULT_GOAL := help

# --- Comandos Principales ---

.PHONY: all help clean build-all build-mac build-windows build-linux all-zip init run

# 🆘 Ayuda
help:
	@echo "📘 \033[1;34mMakefile para SMS/MMS Backup Viewer\033[0m"
	@echo ""
	@echo "Uso: make [comando]"
	@echo ""
	@echo "\033[0;33mComandos de Compilación:\033[0m"
	@echo "  \033[0;32mmake all\033[0m           Compila para TODAS las plataformas."
	@echo "  \033[0;32mmake all-zip\033[0m       Compila todo y genera archivos .zip para distribución."
	@echo ""
	@echo "\033[0;33mCompilación Individual:\033[0m"
	@echo "  \033[0;32mmake build-mac\033[0m     Compila para macOS (Intel y Apple Silicon)."
	@echo "  \033[0;32mmake build-windows\033[0m Compila para Windows (x64)."
	@echo "  \033[0;32mmake build-linux\033[0m   Compila para Linux (x64)."
	@echo ""
	@echo "\033[0;33mUtilidades:\033[0m"
	@echo "  \033[0;32mmake run\033[0m           Ejecuta el código fuente directamente."
	@echo "  \033[0;32mmake clean\033[0m         Elimina las carpetas '$(BUILD_DIR)', '$(DIST_DIR)' y 'Output_Web'."
	@echo "  \033[0;32mmake help\033[0m          Muestra este mensaje de ayuda."
	@echo ""

# 🏗️ Compilar todo
all: clean build-all

build-all: build-mac build-windows build-linux
	@echo "\n✅ \033[1;32m¡Compilación completa!\033[0m Binarios en '$(BUILD_DIR)'"

# 📦 Compilar y Comprimir
all-zip: all
	@echo "\n📦 \033[1;33mComprimiendo binarios...\033[0m"
	@mkdir -p $(DIST_DIR)
	@zip -j $(DIST_DIR)/$(BINARY_NAME)-mac-intel.zip $(BUILD_DIR)/mac_intel/$(BINARY_NAME)
	@zip -j $(DIST_DIR)/$(BINARY_NAME)-mac-arm64.zip $(BUILD_DIR)/mac_arm64/$(BINARY_NAME)
	@zip -j $(DIST_DIR)/$(BINARY_NAME)-linux-x64.zip $(BUILD_DIR)/linux/$(BINARY_NAME)
	@zip -j $(DIST_DIR)/$(BINARY_NAME)-windows-x64.zip $(BUILD_DIR)/windows/$(BINARY_NAME).exe
	@echo "\n🚀 \033[1;32m¡ZIPs listos en la carpeta '$(DIST_DIR)'!\033[0m"

# --- Compilación por Plataforma ---

build-mac:
	@echo "🍎 Compilando para macOS..."
	@mkdir -p $(BUILD_DIR)/mac_intel $(BUILD_DIR)/mac_arm64
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/mac_intel/$(BINARY_NAME) $(SOURCE_FILE)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/mac_arm64/$(BINARY_NAME) $(SOURCE_FILE)

build-windows:
	@echo "🪟 Compilando para Windows..."
	@mkdir -p $(BUILD_DIR)/windows
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(SOURCE_FILE)

build-linux:
	@echo "🐧 Compilando para Linux..."
	@mkdir -p $(BUILD_DIR)/linux
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME) $(SOURCE_FILE)

# --- Utilidades ---

run:
	@go run $(SOURCE_FILE)

clean:
	@echo "🧹 Limpiando archivos..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) Output_Web
	@echo "✨ Sistema limpio."

init:
	@go mod init sms-viewer || true
	@go mod tidy
