package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// --- XML Structs ---

type SMS struct {
	XMLName      xml.Name `xml:"sms"`
	Address      string   `xml:"address,attr"`
	Date         int64    `xml:"date,attr"`
	Type         int      `xml:"type,attr"`
	Body         string   `xml:"body,attr"`
	ContactName  string   `xml:"contact_name,attr"`
	ReadableDate string   `xml:"readable_date,attr"`
}

type MMS struct {
	XMLName      xml.Name `xml:"mms"`
	Address      string   `xml:"address,attr"`
	Date         int64    `xml:"date,attr"`
	MsgBox       int      `xml:"msg_box,attr"` // 1 = Received, 2 = Sent
	ContactName  string   `xml:"contact_name,attr"`
	ReadableDate string   `xml:"readable_date,attr"`
	Parts        struct {
		PartList []Part `xml:"part"`
	} `xml:"parts"`
}

type Part struct {
	Seq  int    `xml:"seq,attr"`
	Ct   string `xml:"ct,attr"`
	Name string `xml:"name,attr"`
	Text string `xml:"text,attr"`
	Data string `xml:"data,attr"`
}

type Call struct {
	XMLName     xml.Name `xml:"call"`
	Number      string   `xml:"number,attr"`
	Duration    int      `xml:"duration,attr"`
	Date        int64    `xml:"date,attr"`
	Type        int      `xml:"type,attr"`
	ContactName string   `xml:"contact_name,attr"`
}

// --- Internal Structs para generar el HTML ---

type Message struct {
	ID          string
	Type        string // "sms", "mms", "call"
	IsIncoming  bool
	Address     string
	ContactName string
	Date        time.Time
	TextBody    string
	MediaFiles  []Media
	Duration    int
}

type Media struct {
	FileName string
	FileType string // "image", "video", "audio", "other"
}

type ContactIndex struct {
	Address     string
	ContactName string
	MsgCount    int
	LastMsgTime time.Time
	FileName    string
}

// Global Paths
var (
	baseOutDir  = "Output_Web"
	mediaOutDir = filepath.Join(baseOutDir, "media")
	htmlOutDir  = filepath.Join(baseOutDir, "chats")
	dataOutDir  = filepath.Join(baseOutDir, "temp_data") // Para no saturar la RAM
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <archivo_backup.xml>")
		os.Exit(1)
	}

	xmlFile := os.Args[1]

	// Crear estructura de directorios
	fmt.Println("[1/3] Preparando carpetas...")
	for _, dir := range []string{baseOutDir, mediaOutDir, htmlOutDir, dataOutDir} {
		os.MkdirAll(dir, os.ModePerm)
	}

	// Fase 1: Parsear XML en Stream y guardar en archivos JSON temporales por contacto
	fmt.Println("[2/3] Parseando archivo XML gigante y extrayendo multimedia (Esto tomará tiempo)...")
	processXMLStream(xmlFile)

	// Fase 2: Generar HTML a partir de los datos procesados
	fmt.Println("[3/3] Generando interfaz web de usuario (HTML)...")
	generateHTML()

	// Limpieza de temporales
	os.RemoveAll(dataOutDir)

	fmt.Printf("\n¡PROCESO COMPLETADO! Abre el archivo '%s' en tu navegador.\n", filepath.Join(baseOutDir, "index.html"))
}

// processXMLStream lee el archivo token por token para no saturar la memoria RAM.
func processXMLStream(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error abriendo archivo: %v", err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	
	// Mapa para mantener abiertos los archivos temporales de cada contacto
	fileMap := make(map[string]*os.File)
	defer func() {
		for _, f := range fileMap {
			f.Close()
		}
	}()

	msgCounter := 0

	for {
		t, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Ignorar errores menores de formato y continuar
			continue
		}

		switch se := t.(type) {
		case xml.StartElement:
			var msg Message
			valid := false

			if se.Name.Local == "sms" {
				var sms SMS
				decoder.DecodeElement(&sms, &se)
				msg = convertSMS(sms)
				valid = true
			} else if se.Name.Local == "mms" {
				var mms MMS
				decoder.DecodeElement(&mms, &se)
				msg = convertMMS(mms, strconv.Itoa(msgCounter))
				valid = true
			} else if se.Name.Local == "call" {
				var call Call
				decoder.DecodeElement(&call, &se)
				msg = convertCall(call)
				valid = true
			}

			if valid {
				saveToTempData(msg, fileMap)
				msgCounter++
				if msgCounter%10000 == 0 {
					fmt.Printf("   ... Procesados %d elementos\n", msgCounter)
				}
			}
		}
	}
	fmt.Printf("Total de elementos procesados: %d\n", msgCounter)
}

func saveToTempData(msg Message, fileMap map[string]*os.File) {
	if msg.Address == "" || msg.Address == "null" {
		msg.Address = "Desconocido"
	}
	
	safeAddr := cleanFileName(msg.Address)
	if safeAddr == "" {
		safeAddr = "Desconocido"
	}

	f, exists := fileMap[safeAddr]
	if !exists {
		path := filepath.Join(dataOutDir, safeAddr+".jsonl")
		var err error
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Error creando temporal para %s: %v", safeAddr, err)
			return
		}
		fileMap[safeAddr] = f
	}

	jsonData, _ := json.Marshal(msg)
	f.Write(jsonData)
	f.WriteString("\n")
}

// --- Limpieza de texto basura y formato ---
func cleanTextBody(text string) string {
	if text == "" || text == "null" {
		return ""
	}
	// 1. Eliminar formato <smil> ... </smil>
	re1 := regexp.MustCompile(`(?is)<smil>.*?</smil>`)
	text = re1.ReplaceAllString(text, "")

	// 2. Eliminar formato &lt;smil&gt; ... &lt;/smil&gt;
	re2 := regexp.MustCompile(`(?is)&lt;smil&gt;.*?&lt;/smil&gt;`)
	text = re2.ReplaceAllString(text, "")

	// 3. Eliminar tabuladores y retornos de carro problemáticos
	text = strings.ReplaceAll(text, "\t", " ") // Convierte tabuladores a espacio simple
	text = strings.ReplaceAll(text, "\r", "")  // Elimina retornos de carro de Windows (\r) dejando solo los \n

	// 4. Colapsar múltiples espacios seguidos en uno solo
	re3 := regexp.MustCompile(` {2,}`)
	text = re3.ReplaceAllString(text, " ")

	// 5. Colapsar saltos de línea excesivos
	re4 := regexp.MustCompile(`\n{3,}`)
	text = re4.ReplaceAllString(text, "\n\n")

	// 6. Limpieza final absoluta de espacios/saltos en los extremos
	return strings.TrimSpace(text)
}

// --- Conversores ---

func convertSMS(sms SMS) Message {
	return Message{
		Type:        "sms",
		IsIncoming:  sms.Type == 1,
		Address:     cleanPhone(sms.Address),
		ContactName: sms.ContactName,
		Date:        time.UnixMilli(sms.Date),
		TextBody:    cleanTextBody(sms.Body),
	}
}

func convertMMS(mms MMS, msgID string) Message {
	msg := Message{
		Type:        "mms",
		IsIncoming:  mms.MsgBox == 1,
		Address:     cleanPhone(mms.Address),
		ContactName: mms.ContactName,
		Date:        time.UnixMilli(mms.Date),
	}

	for i, part := range mms.Parts.PartList {
		if part.Ct == "text/plain" {
			cleanTxt := cleanTextBody(part.Text)
			if cleanTxt != "" {
				msg.TextBody += cleanTxt + "\n"
			}
		}
		if part.Data != "" {
			// Es multimedia en Base64
			decoded, err := base64.StdEncoding.DecodeString(part.Data)
			if err == nil {
				ext := getExtension(part.Ct)
				fileName := fmt.Sprintf("mms_%s_%d%s", msgID, i, ext)
				filePath := filepath.Join(mediaOutDir, fileName)
				os.WriteFile(filePath, decoded, 0644)

				mediaType := "other"
				if strings.HasPrefix(part.Ct, "image/") {
					mediaType = "image"
				} else if strings.HasPrefix(part.Ct, "video/") {
					mediaType = "video"
				} else if strings.HasPrefix(part.Ct, "audio/") {
					mediaType = "audio"
				}

				msg.MediaFiles = append(msg.MediaFiles, Media{
					FileName: fileName,
					FileType: mediaType,
				})
			}
		}
	}
	
	msg.TextBody = strings.TrimSpace(msg.TextBody)
	return msg
}

func convertCall(call Call) Message {
	return Message{
		Type:        "call",
		IsIncoming:  call.Type == 1 || call.Type == 3, // Incoming o Missed
		Address:     cleanPhone(call.Number),
		ContactName: call.ContactName,
		Date:        time.UnixMilli(call.Date),
		Duration:    call.Duration,
	}
}

// --- Utilidades ---

func cleanPhone(phone string) string {
	if phone == "" || phone == "null" {
		return "Desconocido"
	}
	return phone
}

func cleanFileName(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_+]`)
	return reg.ReplaceAllString(name, "")
}

func getExtension(mime string) string {
	switch mime {
	case "image/jpeg", "image/jpg": return ".jpg"
	case "image/png": return ".png"
	case "image/gif": return ".gif"
	case "video/mp4": return ".mp4"
	case "video/3gpp": return ".3gp"
	case "audio/amr": return ".amr"
	case "audio/mp3": return ".mp3"
	default: return ".dat"
	}
}

// --- Generación HTML y UX ---

func generateHTML() {
	files, err := os.ReadDir(dataOutDir)
	if err != nil {
		log.Fatalf("Error leyendo directorio temporal: %v", err)
	}

	var contacts []ContactIndex

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".jsonl") {
			continue
		}

		filePath := filepath.Join(dataOutDir, file.Name())
		messages := readMessagesFromJSONL(filePath)
		
		if len(messages) == 0 {
			continue
		}

		// Ordenar mensajes por fecha cronológica
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Date.Before(messages[j].Date)
		})

		contactName := messages[0].ContactName
		if contactName == "" || contactName == "null" {
			contactName = messages[0].Address
		}
		safeFileName := strings.TrimSuffix(file.Name(), ".jsonl") + ".html"

		// Generar página del chat
		generateChatHTML(contactName, messages, safeFileName)

		// Guardar info para el índice
		contacts = append(contacts, ContactIndex{
			Address:     messages[0].Address,
			ContactName: contactName,
			MsgCount:    len(messages),
			LastMsgTime: messages[len(messages)-1].Date,
			FileName:    "chats/" + safeFileName,
		})
	}

	// Ordenar el índice por el mensaje más reciente
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].LastMsgTime.After(contacts[j].LastMsgTime)
	})

	generateIndexHTML(contacts)
}

func readMessagesFromJSONL(path string) []Message {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var msgs []Message
	scanner := bufio.NewScanner(file)
	// Ampliar el buffer del scanner por si hay mensajes de texto extremadamente largos
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024*10)

	for scanner.Scan() {
		var msg Message
		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err == nil {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

// --- Templates HTML (Diseño UX/UI) ---

const cssStyles = `
<style>
	:root {
		--bg-color: #efeae2;
		--text-color: #111b21;
		--chat-bg-in: #ffffff;
		--chat-bg-out: #d9fdd3;
		--sys-bg: #ffeeba;
		--header-bg: #f0f2f5;
		--meta-color: #667781;
	}
	@media (prefers-color-scheme: dark) {
		:root {
			--bg-color: #0b141a;
			--text-color: #e9edef;
			--chat-bg-in: #202c33;
			--chat-bg-out: #005c4b;
			--sys-bg: #182229;
			--header-bg: #202c33;
			--meta-color: #8696a0;
		}
	}
	body {
		font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
		background-color: var(--bg-color);
		color: var(--text-color);
		margin: 0;
		padding: 0;
	}
	.header {
		background-color: var(--header-bg);
		padding: 15px 20px;
		position: sticky;
		top: 0;
		z-index: 1000;
		box-shadow: 0 1px 3px rgba(0,0,0,0.1);
		display: flex;
		align-items: center;
	}
	.header a { text-decoration: none; color: #007bff; margin-right: 15px; font-weight: bold; }
	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 20px;
		display: flex;
		flex-direction: column;
	}
	.msg {
		max-width: 75%;
		margin-bottom: 4px; 
		padding: 6px 8px 6px 10px; /* Reducido drásticamente para evitar espacios vacíos */
		border-radius: 8px;
		position: relative;
		box-shadow: 0 1px 1px rgba(0,0,0,0.05);
	}
	.msg.in {
		background-color: var(--chat-bg-in);
		align-self: flex-start;
		border-top-left-radius: 0;
	}
	.msg.out {
		background-color: var(--chat-bg-out);
		align-self: flex-end;
		border-top-right-radius: 0;
	}
	.text {
		white-space: pre-wrap; /* Esta regla movida aquí evita que el HTML genere espacios */
		word-wrap: break-word;
		line-height: 1.35;
		margin-bottom: 2px;
		font-size: 0.95em;
	}
	.msg.sys {
		background-color: var(--sys-bg);
		align-self: center;
		text-align: center;
		font-size: 0.85em;
		border-radius: 15px;
		max-width: 90%;
		color: #856404;
		margin-bottom: 12px;
		margin-top: 8px;
		padding: 8px 12px;
	}
	.meta {
		font-size: 0.7em;
		color: var(--meta-color);
		text-align: right;
		line-height: 1;
		margin-top: 0; /* Espacio extra eliminado */
	}
	.media img, .media video {
		max-width: 100%;
		border-radius: 6px;
		margin-top: 5px;
	}
	.media audio { width: 100%; margin-top: 5px; }
	
	/* Index Styles */
	.index-container { max-width: 900px; margin: 20px auto; padding: 0 15px; }
	.contact-card {
		background-color: var(--chat-bg-in);
		padding: 15px;
		margin-bottom: 10px;
		border-radius: 10px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		text-decoration: none;
		color: inherit;
		box-shadow: 0 1px 2px rgba(0,0,0,0.1);
		transition: background-color 0.2s;
	}
	.contact-card:hover { background-color: var(--header-bg); }
	.c-name { font-weight: bold; font-size: 1.1em; margin-bottom: 4px; }
	.c-meta { font-size: 0.85em; color: var(--meta-color); }
	.c-badge { background-color: #25d366; color: white; padding: 4px 10px; border-radius: 20px; font-size: 0.8em; font-weight: bold;}
</style>
`

func generateChatHTML(contactName string, messages []Message, fileName string) {
	tmplStr := `<!DOCTYPE html>
<html lang="es">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Chat con {{.ContactName}}</title>
	` + cssStyles + `
</head>
<body>
	<div class="header">
		<a href="../index.html">← Volver</a>
		<h2>{{.ContactName}}</h2>
	</div>
	<div class="container">
		{{range .Messages}}
			{{if eq .Type "call"}}
				<div class="msg sys">
					📞 Llamada {{if .IsIncoming}}Entrante/Perdida{{else}}Saliente{{end}} - {{.Duration}} segundos
					<div class="meta">{{.Date.Format "02/01/2006 15:04"}}</div>
				</div>
			{{else}}
				<div class="msg {{if .IsIncoming}}in{{else}}out{{end}}">
					{{if .TextBody}}<div class="text">{{.TextBody}}</div>{{end}}
					{{if .MediaFiles}}
						<div class="media">
							{{range .MediaFiles}}
								{{if eq .FileType "image"}} <img src="../media/{{.FileName}}" loading="lazy" />
								{{else if eq .FileType "video"}} <video controls src="../media/{{.FileName}}"></video>
								{{else if eq .FileType "audio"}} <audio controls src="../media/{{.FileName}}"></audio>
								{{else}} <a href="../media/{{.FileName}}" target="_blank">📁 Archivo Adjunto</a>
								{{end}}
							{{end}}
						</div>
					{{end}}
					<div class="meta">{{.Date.Format "02/01/2006 15:04"}}</div>
				</div>
			{{end}}
		{{end}}
	</div>
</body>
</html>`

	tmpl, err := template.New("chat").Parse(tmplStr)
	if err != nil {
		log.Printf("Error parseando template: %v", err)
		return
	}

	outPath := filepath.Join(htmlOutDir, fileName)
	file, err := os.Create(outPath)
	if err != nil {
		log.Printf("Error creando archivo html: %v", err)
		return
	}
	defer file.Close()

	data := struct {
		ContactName string
		Messages    []Message
	}{
		ContactName: contactName,
		Messages:    messages,
	}

	tmpl.Execute(file, data)
}

func generateIndexHTML(contacts []ContactIndex) {
	tmplStr := `<!DOCTYPE html>
<html lang="es">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Copia de Seguridad SMS/MMS</title>
	` + cssStyles + `
</head>
<body>
	<div class="header" style="justify-content: center;">
		<h2>Copia de Seguridad SMS / MMS / Llamadas</h2>
	</div>
	<div class="index-container">
		{{range .}}
		<a href="{{.FileName}}" class="contact-card">
			<div>
				<div class="c-name">{{.ContactName}}</div>
				<div class="c-meta">{{.Address}} • Último msj: {{.LastMsgTime.Format "02/01/2006"}}</div>
			</div>
			<div class="c-badge">{{.MsgCount}} msjs</div>
		</a>
		{{end}}
	</div>
</body>
</html>`

	tmpl, err := template.New("index").Parse(tmplStr)
	if err != nil {
		log.Fatalf("Error parseando template index: %v", err)
	}

	outPath := filepath.Join(baseOutDir, "index.html")
	file, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("Error creando index.html: %v", err)
	}
	defer file.Close()

	tmpl.Execute(file, contacts)
}
