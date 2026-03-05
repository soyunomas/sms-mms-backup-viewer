# SMS / MMS Backup Viewer 📱💬

An ultra-fast and efficient tool written in **Go** to convert massive SMS, MMS, and Call Log backups (in XML format) into a beautiful local web interface (HTML/CSS) styled like WhatsApp or iMessage.

Developed to process massive files (tested with backups over 5GB) without crashing your computer's RAM.

## ✨ Features

- **Stream Parsing:** Reads the XML file sequentially. Forget about *Out of Memory* errors.
- **Multimedia Extraction:** Detects and interprets Base64 code embedded in MMS, extracting images, videos, and audio directly to a local folder.
- **Smart Cleaning:** Automatically removes technical junk from messages (such as `<smil>` tags, tabs, or extreme empty spaces).
- **Responsive UI/UX:** Generates static HTML files with a modern chat design.
- **Auto Dark/Light Mode:** The design automatically respects your operating system's theme settings.
- **Total Privacy:** Everything runs locally. Your messages are never uploaded to any server.

---

## 🚀 Getting Started (Quick Download)

You don't need to install Go or know how to code to use this tool.

1. Go to the **[Releases](https://github.com/soyunomas/sms-mms-backup-viewer/releases)** section of this repository.
2. Download the `.zip` file for your operating system (Windows, macOS, or Linux).
3. Extract the file to a folder on your computer.
4. Place your backup file (e.g., `backup.xml`) in that same folder.
5. Open a terminal or command prompt in that folder and run:
   - **Windows:** `sms-viewer.exe backup.xml`
   - **macOS / Linux:** `./sms-viewer backup.xml`
6. Once finished, open the `Output_Web/index.html` file with your favorite web browser.

---

## 🛠️ Developer Installation

If you prefer to run it from source or compile it yourself:

### Prerequisites
- [Go (Golang)](https://go.dev/dl/) installed.
- (Optional) `make` to use the provided Makefile.

### Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/soyunomas/sms-mms-backup-viewer.git
   cd sms-mms-backup-viewer
   ```

2. Run directly:
   ```bash
   go run main.go backup.xml
   ```

3. Or compile using the Makefile:
   ```bash
   make help      # View build options
   make all-zip   # Generate all compressed binaries in /dist
   ```

---

## 📁 Output Structure

Upon completion, the application will create a folder named `Output_Web` with the following structure:

```text
Output_Web/
 ├── index.html       # Main index with all your contacts
 ├── chats/           # Folder containing individual conversations (.html)
 └── media/           # Folder containing all extracted photos, audio, and videos
```

## 📝 License

This project is licensed under the MIT License. Feel free to use, modify, and share it.


