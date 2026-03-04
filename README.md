# SMS / MMS Backup Viewer 📱💬

Una herramienta ultrarrápida y eficiente escrita en **Go** para convertir copias de seguridad gigantescas de SMS, MMS y Registro de Llamadas (en formato XML) en una atractiva interfaz web local (HTML/CSS) estilo WhatsApp/iMessage.

Desarrollada para procesar archivos masivos (probado con copias de seguridad de más de 5GB) sin colapsar la memoria RAM de tu ordenador.

## ✨ Características

- **Análisis por Flujos (Stream Parsing):** Lee el archivo XML secuencialmente. Olvídate de los errores de *Out of Memory*.
- **Extracción de Multimedia:** Detecta e interpreta código Base64 incrustado en los MMS extrayendo imágenes, vídeos y audios directamente a una carpeta local.
- **Limpieza Inteligente:** Elimina automáticamente basura técnica de los mensajes (como etiquetas `<smil>`, tabuladores o espacios vacíos extremos).
- **Interfaz UI/UX Responsiva:** Genera archivos HTML estáticos con un diseño tipo chat moderno.
- **Modo Oscuro/Claro Automático:** El diseño respeta la configuración de tema de tu sistema operativo.
- **Privacidad Total:** Todo se ejecuta en local. Tus mensajes no se suben a ningún servidor.

## 🚀 Requisitos previos

- Tener instalado [Go (Golang)](https://go.dev/dl/).

## 🛠️ Instalación y Uso

1. Clona este repositorio en tu ordenador:
   ```bash
   git clone https://github.com/soyunomas/sms-mms-backup-viewer.git
   cd sms-mms-backup-viewer
   ```

2. Coloca tu archivo de copia de seguridad (ej. `backup.xml`) en la misma carpeta.

3. Ejecuta el script pasándole el nombre de tu archivo como argumento:
   ```bash
   go run main.go backup.xml
   ```

4. **¡Espera!** Si tu archivo es de varios Gigabytes, el proceso tomará un tiempo. Podrás ver el progreso en la consola.
5. Al finalizar, abre el archivo `Output_Web/index.html` con tu navegador web favorito.

## 📁 Estructura de salida

Al terminar la ejecución, la aplicación creará una carpeta llamada `Output_Web` con esta estructura:

```text
Output_Web/
 ├── index.html       # Índice principal con todos tus contactos
 ├── chats/           # Carpeta con las conversaciones individuales (.html)
 └── media/           # Carpeta con todas las fotos, audios y vídeos extraídos
```

## 📝 Licencia

Este proyecto está bajo la Licencia MIT 


