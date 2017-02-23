package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//------------------------------------------------------------------------------
func check_ResponseToHTTP(err error, w http.ResponseWriter) {
	if err != nil {
		fmt.Fprintln(w, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}

//alle Templates die im Ordner sind werden geParset
var t = template.Must(template.ParseGlob("daten/tpl/*"))

//var server = "mongodb://borsti.inf.fh-flensburg.de:27017"

var server = "localhost"

var dbSession, err = mgo.Dial(server)
var db = dbSession.DB("HA15DB_Christian_Schulz_570249")

// damit der Lade Bildschirm die Prozentzahl hat
var prozentAnzeigeZahl int

//Collection wählen/erstellen
var userPW = db.C("namePW")
var pools = db.C("pools")

//structs ineinander gebaut
type prozentStatusTyp struct {
	Zufall  int
	Prozent int
}
type userTyp struct {
	Name     string
	Passwort string
}
type poolEigenschaftenTyp struct {
	Anzahl        int
	KachelGroesse int
	Helligkeit    int
	Farbverlauf   bool
	HistoBild     string
}
type bilderTyp struct {
	Bild              string
	MittlererFarbWert [3]uint32
}

type sammlungBilderTyp struct {
	BildName   string
	BildNameDB string
	Hohe       int
	Breite     int
	HistoBild  string
	ThumbBild  string
}
type mosaikBilderTyp struct {
	MosaikName         string
	MosaikNameDB       string
	MosaikHoehe        int
	MosaikBreite       int
	BasisBild          sammlungBilderTyp
	PoolEigenschaften  poolTyp
	DoppelteVerwendung bool
	NbestGeeignet      int
	ThumbBild          string
	SammlungName       string
	//wenn 0 dann ist es die best geeignete
}
type poolTyp struct {
	Name          string
	GenBilder     bool
	Eigenschaften poolEigenschaftenTyp
	Bilder        []bilderTyp
}
type sammlungTyp struct {
	Name   string
	Bilder []sammlungBilderTyp
}
type mosaikTyp struct {
	Bilder []mosaikBilderTyp
}
type poolsTyp struct {
	Name           string
	Pools          []poolTyp
	Sammlung       []sammlungTyp
	MosaikSammlung mosaikTyp
}
type passendeKachelTyp struct {
	Name      string
	KleinsteD float64
}

type HSL struct {
	H, S, L float64
}

//Sortier Funktionenen
type SortByKleinstesD []passendeKachelTyp

func (a SortByKleinstesD) Len() int      { return len(a) }
func (a SortByKleinstesD) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortByKleinstesD) Less(i, j int) bool {
	itVarI := a[i].KleinsteD
	itVarJ := a[j].KleinsteD

	return itVarI < itVarJ
}

//Löscht alle Bilder aus dem Grid FS des Users und den eintrag des Users in beiden Collections
func delete(w http.ResponseWriter, r *http.Request) {

	keks, _ := r.Cookie("user")
	//**********************Gridfs löschen*****************************
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	prozentAnzeigeZahl = 0

	for i := 0; i < len(ergDatenLog.Pools); i++ {
		gridfs.Remove(ergDatenLog.Pools[i].Eigenschaften.HistoBild)
		for j := 0; j < len(ergDatenLog.Pools[i].Bilder); j++ {
			gridfs.Remove(ergDatenLog.Pools[i].Bilder[j].Bild)
		}
	}

	prozentAnzeigeZahl = 33
	for i := 0; i < len(ergDatenLog.Sammlung); i++ {
		for j := 0; j < len(ergDatenLog.Sammlung[i].Bilder); j++ {
			gridfs.Remove(ergDatenLog.Sammlung[i].Bilder[j].BildNameDB)
			gridfs.Remove(ergDatenLog.Sammlung[i].Bilder[j].ThumbBild)
			gridfs.Remove(ergDatenLog.Sammlung[i].Bilder[j].HistoBild)

		}
	}
	prozentAnzeigeZahl = 66
	for i := 0; i < len(ergDatenLog.MosaikSammlung.Bilder); i++ {
		gridfs.Remove(ergDatenLog.MosaikSammlung.Bilder[i].MosaikNameDB)
		gridfs.Remove(ergDatenLog.MosaikSammlung.Bilder[i].ThumbBild)

	}
	prozentAnzeigeZahl = 99
	//************************************************************
	userPW.Remove(bson.M{"name": bson.M{"$eq": keks.Value}})
	pools.Remove(bson.M{"name": bson.M{"$eq": keks.Value}})
	newCookie := http.Cookie{Name: "user", MaxAge: -1}
	newCookie2 := http.Cookie{Name: "aktivePool", MaxAge: -1}
	newCookie3 := http.Cookie{Name: "aktiveSammlung", MaxAge: -1}
	http.SetCookie(w, &newCookie)
	http.SetCookie(w, &newCookie2)
	http.SetCookie(w, &newCookie3)
	prozentAnzeigeZahl = 0
}

//Login aus der Übung aber abgeändert
func loginSeite(w http.ResponseWriter, r *http.Request) {
	check := func(err error) {
		if err != nil {
			fmt.Println(err)
		}
	}
	keks, err := r.Cookie("user")
	check(err)
	if keks != nil {
		tempName := keks.Value
		t.ExecuteTemplate(w, "tessera.html", tempName)
	} else {
		switch r.Method {
		case "GET":
			t.ExecuteTemplate(w, "login.html", nil)

		case "POST":
			r.ParseForm()
			name := r.FormValue("name")
			pw1 := r.FormValue("pw1")

			var ergDatenLog userTyp
			var err = userPW.Find(bson.M{"name": bson.M{"$eq": name}}).One(&ergDatenLog) // entspricht findOne
			check(err)
			//Sicherheitsabfrage
			if name == "" {
				t.ExecuteTemplate(w, "login.html", "Bitte Name eingeben")
			} else if pw1 == "" {
				t.ExecuteTemplate(w, "login.html", "Bitte Passwort eingeben")
			} else if ergDatenLog.Name == name && ergDatenLog.Passwort == pw1 {
				newCookie := http.Cookie{Name: "user", Value: name}
				http.SetCookie(w, &newCookie)
				tempName := name
				t.ExecuteTemplate(w, "tessera.html", tempName)
			} else {
				t.ExecuteTemplate(w, "login.html", "falsches Passwort")
			}
		}
	}
}

//hier werden einfach nur die Cookies gelöscht
func logout(w http.ResponseWriter, r *http.Request) {

	newCookie := http.Cookie{Name: "user", MaxAge: -1}
	newCookie2 := http.Cookie{Name: "aktivePool", MaxAge: -1}
	newCookie3 := http.Cookie{Name: "aktiveSammlung", MaxAge: -1}
	http.SetCookie(w, &newCookie)
	http.SetCookie(w, &newCookie2)
	http.SetCookie(w, &newCookie3)

}

//Upload für Pool Bilder
func multipartFormHandler(w http.ResponseWriter, r *http.Request) {
	//Temp oderner erstellen
	os.Mkdir("./temps", 0700)
	prozentAnzeigeZahl = 0
	var bilderGrose int
	bilderGrose = 100
	keks, err := r.Cookie("user")
	poolKeks, err := r.Cookie("aktivePool")
	poolname := poolKeks.Value
	user := keks.Value
	// GridFs-collection "bilder" dieser DB:
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)

	//***********************Anker damit Bilder den User/Pools zugeordnet werden können *****************

	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog) // entspricht findOne

	for j := 0; j < len(ergDatenLog.Pools); j++ {
		if poolKeks.Value == ergDatenLog.Pools[j].Name {
			bilderGrose = ergDatenLog.Pools[j].Eigenschaften.KachelGroesse
		}
	}

	err = r.ParseMultipartForm(2000000) // bytes

	if err == nil { // => alles ok
		formdataZeiger := r.MultipartForm
		if formdataZeiger != nil { // beim ersten request ist die Form leer!
			//Holt sich alle files die er bekommt
			files := formdataZeiger.File["hochladeInputFileTagName"]

			anzahlFiles := len(files)
			count := -1
			for i, _ := range files {
				count++
				prozentAnzeigeZahl = (count * 100) / anzahlFiles
				//*****************Namensfindung***ohneSpeicherungOrginal**********

				erg := make(map[string]interface{})
				iter := gridfs.Find(nil).Iter()
				zahl := 1
				for iter.Next(erg) {

					for erg["filename"] == files[i].Filename {
						files[i].Filename = strconv.Itoa(zahl) + files[i].Filename
						zahl++
					}
				}
				//*******************ende Namesfindung

				// upload-files öffnen:
				uplFile, err := files[i].Open()
				check_ResponseToHTTP(err, w)

				//neues File im Temp Oderner erzeugen
				tempBildRoh, _ := os.Create("./temps/tempBildRoh.jpg")

				_, err = io.Copy(tempBildRoh, uplFile)

				//Bild aus dem Temp öffnen
				imgRoh, err := os.Open("./temps/tempBildRoh.jpg")
				check_ResponseToHTTP(err, w)

				//bilder schneiden
				imgTemp, _, _ := image.Decode(imgRoh)

				//Bilder Größtes Viereck wählen
				var hohe = imgTemp.Bounds().Size().Y
				var breite = imgTemp.Bounds().Size().X
				var startX int
				var startY int
				var size int
				if breite <= hohe {
					startX = 0
					startY = (hohe - breite) / 2
					size = breite

				} else {
					startY = 0
					startX = (breite - hohe) / 2
					size = hohe
				}
				// neues Bild erzeugen
				newPic := image.NewRGBA(image.Rect(0, 0, size, size))
				//malt Inhalt in das neue Bild
				draw.Draw(newPic, newPic.Bounds(), imgTemp, image.Point{startX, startY}, draw.Src)

				//neues File im Temp Ordner erzeugen
				tempBildQuadrat, _ := os.Create("./temps/tempBildQuadrat.jpg")
				//mache es zu eine jpeg
				jpeg.Encode(tempBildQuadrat, newPic, &jpeg.Options{jpeg.DefaultQuality})

				//Bild aus dem Temp öffnen
				tempBildFertig, err := imaging.Open("./temps/tempBildQuadrat.jpg")

				img := imaging.Resize(tempBildFertig, bilderGrose, bilderGrose, imaging.CatmullRom)
				err = imaging.Save(img, "./temps/tempBildFertig.jpg")

				//******************MittlererFarbWert*********************
				var r uint32
				var g uint32
				var b uint32
				r, g, b = 0, 0, 0
				mittleWertImg, _ := imaging.Open("./temps/tempBildFertig.jpg")
				for i := 0; i < mittleWertImg.Bounds().Size().Y; i++ {
					for j := 0; j < mittleWertImg.Bounds().Size().X; j++ {
						rTemp, gTemp, bTemp, _ := mittleWertImg.At(j, i).RGBA()
						r, g, b = r+uint32(rTemp), g+uint32(gTemp), b+uint32(bTemp)
					}
				}
				allePixels := uint32(mittleWertImg.Bounds().Max.X * mittleWertImg.Bounds().Max.Y)

				mittlereWert := [3]uint32{r / allePixels, g / allePixels, b / allePixels}
				//********************************************************

				//Bild aus dem Temp öffnen
				imgFertig, err := os.Open("./temps/tempBildFertig.jpg")
				check_ResponseToHTTP(err, w)

				//***********************Anker damit Bilder den User/Pools zugeordnet werden können *****************

				//hochgeladene Bild auch in der COllection mit Name versehen
				for j := 0; j < len(ergDatenLog.Pools); j++ {
					if poolKeks.Value == ergDatenLog.Pools[j].Name {

						ergDatenLog.Pools[j].Eigenschaften.Anzahl++
						ergDatenLog.Pools[j].Bilder = append(ergDatenLog.Pools[j].Bilder, bilderTyp{Bild: files[i].Filename, MittlererFarbWert: mittlereWert})
						pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, ergDatenLog)

					}
				}
				//*****************************************************************************************************

				// grid-file mit diesem Namen erzeugen:
				gridFile, err := gridfs.Create(files[i].Filename)
				check_ResponseToHTTP(err, w)

				// in GridFSkopieren:
				_, err = io.Copy(gridFile, imgFertig)
				check_ResponseToHTTP(err, w)

				//schliesen der geöffneten sachen
				err = gridFile.Close()
				check_ResponseToHTTP(err, w)
				imgFertig.Close()
				tempBildRoh.Close()
				uplFile.Close()
				imgRoh.Close()
				tempBildQuadrat.Close()
				imgRoh.Close()
			}

		}
		prozentAnzeigeZahl = 99
		macheHistoBildPool(poolname, user)

	}
	//Löschen des temps ordener
	os.RemoveAll("./temps/")
	prozentAnzeigeZahl = 0
}

//Basis Bilder Upload
func uploadBilderSammlung(w http.ResponseWriter, r *http.Request) {
	os.Mkdir("./temps", 0700)
	keks, _ := r.Cookie("user")
	aktiveSammlung, _ := r.Cookie("aktiveSammlung")
	var bilderGrose int
	bilderGrose = 150
	// GridFs-collection "bilder" dieser DB:
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)

	err = r.ParseMultipartForm(2000000) // bytes
	prozentAnzeigeZahl = 0
	if err == nil { // => alles ok
		formdataZeiger := r.MultipartForm

		if formdataZeiger != nil { // beim ersten request ist die Form leer!
			files := formdataZeiger.File["hochladeInputFileTagName"]

			//berchnungen für den lade Bildschirm
			anzahlFiles := len(files)
			count := 0

			for i, _ := range files {
				count++
				prozentAnzeigeZahl = (count * 100) / anzahlFiles

				// upload-files öffnen:
				uplFile, err := files[i].Open()
				check_ResponseToHTTP(err, w)

				//*****************Namensfindung*****************
				originalBildName := files[i].Filename
				erg := make(map[string]interface{})
				iter := gridfs.Find(nil).Iter()
				zahl := 1
				for iter.Next(erg) {

					for erg["filename"] == files[i].Filename {
						files[i].Filename = strconv.Itoa(zahl) + files[i].Filename
						zahl++
					}
				}

				//neues File im Temp Oderner erzeugen
				tempBildRoh, _ := os.Create("./temps/tempBildRoh.jpg")

				_, err = io.Copy(tempBildRoh, uplFile)

				//Bild aus dem Temp öffnen
				imgRoh, err := os.Open("./temps/tempBildRoh.jpg")
				check_ResponseToHTTP(err, w)
				//bilder schneiden
				imgTemp, _, _ := image.Decode(imgRoh)

				tempHohe := imgTemp.Bounds().Size().Y
				tempBreite := imgTemp.Bounds().Size().X

				//neues File im Temp Ordner erzeugen
				tempBildQuadrat, _ := os.Create("./temps/tempBildQuadrat.jpg")
				//wird zum JPEG gemacht
				jpeg.Encode(tempBildQuadrat, imgTemp, &jpeg.Options{jpeg.DefaultQuality})

				//*************HistoBild******************
				histoFileName := "histo_" + files[i].Filename
				machHistoBildSammlung()

				//****************************************

				//Bild aus dem Temp öffnen
				tempBildFertig, err := imaging.Open("./temps/tempBildQuadrat.jpg")

				img := imaging.Thumbnail(tempBildFertig, bilderGrose, bilderGrose, imaging.CatmullRom)
				err = imaging.Save(img, "./temps/tempBildFertig.jpg")

				//Bild aus dem Temp öffnen
				imgFertig, err := os.Open("./temps/tempBildFertig.jpg")
				check_ResponseToHTTP(err, w)

				//***********************Anker damit Bilder den User/Sammlungen zugeordnet werden können *****************

				var ergDatenLog poolsTyp
				pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog) // entspricht findOne

				//hochgeladene Bild auch in der COllection mit Name versehen
				for j := 0; j < len(ergDatenLog.Sammlung); j++ {

					if aktiveSammlung.Value == ergDatenLog.Sammlung[j].Name {
						ergDatenLog.Sammlung[j].Bilder = append(ergDatenLog.Sammlung[j].Bilder,
							sammlungBilderTyp{BildName: originalBildName, BildNameDB: files[i].Filename, Hohe: tempHohe, Breite: tempBreite, HistoBild: histoFileName, ThumbBild: "thumbBild" + files[i].Filename})
						pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, ergDatenLog)

					}
				}
				//*****************************************************************************************************
				//Bild aus dem Temp öffnen
				orgiBild, err := os.Open("./temps/tempBildRoh.jpg")
				check_ResponseToHTTP(err, w)

				bildHisto, _ := os.Open("./temps/tempBildHisto.jpg")

				// grid-file mit diesem Namen erzeugen:
				gridFileHisto, err := gridfs.Create(histoFileName)
				check_ResponseToHTTP(err, w)

				// in GridFSkopieren:
				_, _ = io.Copy(gridFileHisto, bildHisto)
				_ = gridFileHisto.Close()

				// grid-file mit diesem Namen erzeugen:
				gridFile, err := gridfs.Create(files[i].Filename)
				check_ResponseToHTTP(err, w)

				// in GridFSkopieren:
				_, err = io.Copy(gridFile, orgiBild)
				check_ResponseToHTTP(err, w)

				err = gridFile.Close()
				check_ResponseToHTTP(err, w)

				// grid-file mit diesem Namen erzeugen:
				gridFileThumb, err := gridfs.Create("thumbBild" + files[i].Filename)
				check_ResponseToHTTP(err, w)

				// in GridFSkopieren:
				_, err = io.Copy(gridFileThumb, imgFertig)
				check_ResponseToHTTP(err, w)

				err = gridFileThumb.Close()
				check_ResponseToHTTP(err, w)

				//schliesen der geöffneten temp Ordner Datein
				imgFertig.Close()
				uplFile.Close()
				bildHisto.Close()
				orgiBild.Close()
				imgRoh.Close()
				tempBildQuadrat.Close()
				imgRoh.Close()
				tempBildRoh.Close()
			}
			prozentAnzeigeZahl = 0
		}

	}

	os.RemoveAll("./temps/")
}

//********Histogramm Funktionene **********
//macht aus RGB Werten HSL Werte
func farbumwandlungzuHSL(r, g, b uint8) (intensity, saturation, light float64) {
	//teilt die Farb Werte durch 255
	farbwertRot := float64(r) / 255
	farbwertGruen := float64(g) / 255
	farbwertBlau := float64(b) / 255

	max := math.Max(math.Max(farbwertRot, farbwertGruen), farbwertBlau)
	min := math.Min(math.Min(farbwertRot, farbwertGruen), farbwertBlau)
	light = (max + min) / 2
	if max == min {
		intensity, saturation = 0, 0
	} else {

		d := max - min
		if light > 0.5 {
			saturation = d / (2.0 - max - min)
		} else {
			saturation = d / (max + min)
		}
		switch max {
		case farbwertRot:
			intensity = (farbwertGruen - farbwertBlau) / d
			if farbwertGruen < farbwertBlau {
				intensity += 6
			}
		case farbwertGruen:
			intensity = (farbwertBlau-farbwertRot)/d + 2
		case farbwertBlau:
			intensity = (farbwertRot-farbwertGruen)/d + 4
		}
		intensity /= 6
	}
	return
}

// von HSL ind RGB werte
func farbumwandlungzuRGB(intensity, saturation, light float64) (r, g, b uint8) {
	var farbwertRot, farbwertGruen, farbwertBlau float64
	if saturation == 0 {
		farbwertRot, farbwertGruen, farbwertBlau = light, light, light
	} else {
		var q float64
		if light < 0.5 {
			q = light * (1 + saturation)
		} else {
			q = light + saturation - saturation*light
		}
		p := 2*light - q
		farbwertRot = farbwechsel(p, q, intensity+1.0/3)
		farbwertGruen = farbwechsel(p, q, intensity)
		farbwertBlau = farbwechsel(p, q, intensity-1.0/3)
	}
	r = uint8((farbwertRot * 255) + 0.5)
	g = uint8((farbwertGruen * 255) + 0.5)
	b = uint8((farbwertBlau * 255) + 0.5)
	return
}
func farbwechsel(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6 {
		return p + (q-p)*6*t
	}
	if t < 0.5 {
		return q
	}
	if t < 2.0/3 {
		return p + (q-p)*(2.0/3-t)*6
	}
	return p
}

//******************
func machHistoBildSammlung() {
	imgRoh, _ := imaging.Open("./temps/tempBildQuadrat.jpg")
	//Farben in dennen die Helligkeit und die Farbverteilug angezeigt werden sollen
	yellowColor := color.NRGBA{255, 228, 0, 255}
	wihteColor := color.NRGBA{255, 255, 255, 255}
	var werte []HSL

	for i := 0; i < imgRoh.Bounds().Size().Y; i++ {
		rTemp, gTemp, bTemp, _ := imgRoh.At(0, i).RGBA()
		h, s, l := farbumwandlungzuHSL(uint8(rTemp>>8), uint8(gTemp>>8), uint8(bTemp>>8))
		for j := 1; j < imgRoh.Bounds().Size().X; j++ {
			rTemp, gTemp, bTemp, _ := imgRoh.At(j, i).RGBA()
			//Wandelt die RGB in HSL werte um
			hNeu, sNeu, lNeu := farbumwandlungzuHSL(uint8(rTemp>>8), uint8(gTemp>>8), uint8(bTemp>>8))
			//addiert die HSL werte drauf
			h = h + hNeu
			s = s + sNeu
			l = l + lNeu
		}

		h = h / float64(imgRoh.Bounds().Size().X)
		s = s / float64(imgRoh.Bounds().Size().X)
		l = l / float64(imgRoh.Bounds().Size().X)
		//werte werden in das array gespeichert
		werte = append(werte, HSL{h, s, l})
	}
	breiteFarbPallete := 10
	imgHisto := image.NewRGBA(image.Rect(0, 0, len(werte)+breiteFarbPallete, 150))

	// FarbPallete Werte und male die Farbpallette
	for y := 0; y < imgHisto.Bounds().Size().Y; y++ {
		h := float64(imgHisto.Bounds().Size().Y-y) / float64(imgHisto.Bounds().Size().Y)
		r, g, b := farbumwandlungzuRGB(h, 1, 0.5)
		color := color.NRGBA{r, g, b, 255}
		for x := 0; x < breiteFarbPallete; x++ {
			imgHisto.Set(x, y, color)
		}
	}
	//Das wirklcihe Histogramm malen
	for x := breiteFarbPallete; x < imgHisto.Bounds().Size().X; x++ {
		for y := 0; y < imgHisto.Bounds().Size().Y; y++ {
			//Zeichne H-Wert
			if y == imgHisto.Bounds().Size().Y-int(werte[x-breiteFarbPallete].H*float64(imgHisto.Bounds().Size().Y)) {
				imgHisto.Set(x, y, yellowColor)
			}
			if y == imgHisto.Bounds().Size().Y-int(werte[x-breiteFarbPallete].L*float64(imgHisto.Bounds().Size().Y)) {
				imgHisto.Set(x, y, wihteColor)
			}
		}
	}

	tempBildHisto, _ := os.Create("./temps/tempBildHisto.jpg")
	//bild als peg erstellene
	jpeg.Encode(tempBildHisto, imgHisto, &jpeg.Options{jpeg.DefaultQuality})
	tempBildHisto.Close()
}
func macheHistoBildPool(poolName string, user string) {
	//Braucht das Grid FS um jedes bild einzulesen
	gridfs := db.GridFS("bilder")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": user}}).One(&ergDatenLog) // entspricht findOne
	//aktuellen Pool finden
	var tempPool poolTyp
	for j := 0; j < len(ergDatenLog.Pools); j++ {
		if poolName == ergDatenLog.Pools[j].Name {
			tempPool = ergDatenLog.Pools[j]

		}
	}
	//ganzen Pool durch gehen und jedes Bild bearbeitetn
	if tempPool.Eigenschaften.Anzahl > 0 {
		//Erstellt Farben für das HistoBild der Pools
		yellowColor := color.NRGBA{255, 228, 0, 255}
		wihteColor := color.NRGBA{255, 255, 255, 255}
		// HSL Struct
		var werteEinBild []HSL
		//Kachelbreite
		breiteBild := tempPool.Eigenschaften.KachelGroesse
		//HSL Float Variablen
		var h, s, l float64

		for k := 0; k < tempPool.Eigenschaften.Anzahl; k++ {

			gridfsName := "bilder"
			gridfs := db.GridFS(gridfsName)

			// file aus GridFS lesen und als response senden:
			gridFile, _ := gridfs.Open(tempPool.Bilder[k].Bild)
			//neues File im Temp Oderner erzeugen
			tempBildRoh, _ := os.Create("./temps/tempBildRoh.jpg")
			//Kopiert org Bild in TempBild
			io.Copy(tempBildRoh, gridFile)
			gridFile.Close()
			tempBildRoh.Close()
			imgRoh, _ := imaging.Open("./temps/tempBildRoh.jpg")
			for i := 0; i < breiteBild; i++ {
				for j := 0; j < breiteBild; j++ {
					//Gibt RGBA Wert bei dem Pixel aus
					rTemp, gTemp, bTemp, _ := imgRoh.At(j, i).RGBA()
					// Wandelt RGBA Werte zu HSL werten.
					hNeu, sNeu, lNeu := farbumwandlungzuHSL(uint8(rTemp>>8), uint8(gTemp>>8), uint8(bTemp>>8))
					h = h + hNeu
					s = s + sNeu
					l = l + lNeu
				}
			}
			//mittewert
			h = h / float64(breiteBild*breiteBild)
			s = s / float64(breiteBild*breiteBild)
			l = l / float64(breiteBild*breiteBild)
			werteEinBild = append(werteEinBild, HSL{h, s, l})
		}
		//x-achse Von Histo => werteEinBild-Laenge

		breiteFarbPallete := 10
		//breite mal einen Faktor
		imgHisto := image.NewRGBA(image.Rect(0, 0, len(werteEinBild)+breiteFarbPallete, 150))

		// Farbplette malen
		for y := 0; y < imgHisto.Bounds().Size().Y; y++ {
			h := float64(imgHisto.Bounds().Size().Y-y) / float64(imgHisto.Bounds().Size().Y)
			r, g, b := farbumwandlungzuRGB(h, 1, 0.5)
			color := color.NRGBA{r, g, b, 255}

			for x := 0; x < breiteFarbPallete; x++ {
				imgHisto.Set(x, y, color)
			}
		}
		//Historgramm malen Punkte/Linien Zeichen
		for x := breiteFarbPallete; x < imgHisto.Bounds().Size().X; x++ {
			for y := 0; y < imgHisto.Bounds().Size().Y; y++ {
				//Zeichne H-Wert(Linien) in Histo IMG Farbwerte
				if y == imgHisto.Bounds().Size().Y-int(werteEinBild[x-breiteFarbPallete].H*float64(imgHisto.Bounds().Size().Y)) {
					imgHisto.Set(x, y, yellowColor)
				}
				//Zeichne L-Wert(Linien) in Histo IMG Helligkeit
				if y == imgHisto.Bounds().Size().Y-int(werteEinBild[x-breiteFarbPallete].L*float64(imgHisto.Bounds().Size().Y)) {
					imgHisto.Set(x, y, wihteColor)
				}
			}
		}

		tempBildHisto, _ := os.Create("./temps/tempBildHisto.jpg")

		jpeg.Encode(tempBildHisto, imgHisto, &jpeg.Options{jpeg.DefaultQuality})
		tempBildHisto.Close()
	} else {
		imgHisto := image.NewRGBA(image.Rect(0, 0, 200, 200))
		tempBildHisto, _ := os.Create("./temps/tempBildHisto.jpg")
		jpeg.Encode(tempBildHisto, imgHisto, &jpeg.Options{jpeg.DefaultQuality})
	}
	fertigHisto, _ := os.Open("./temps/tempBildHisto.jpg")
	histBildString := "histoBild_" + poolName + user
	gridfs.Remove(histBildString)
	gridFile, _ := gridfs.Create(histBildString)
	// Kopiert Histo in GridBild
	io.Copy(gridFile, fertigHisto)
	fertigHisto.Close()

	gridFile.Close()
}

//Holt das Image aus dem Grid und zeigt es an
func getImageHandler(w http.ResponseWriter, r *http.Request) {

	// request lesen:
	r_dbName := r.URL.Query().Get("dbName")
	r_gridfsName := r.URL.Query().Get("gridfsName")
	r_fileName := r.URL.Query().Get("fileName")

	// DB-Verbindung:
	session, err := mgo.Dial(server)
	check_ResponseToHTTP(err, w)
	defer session.Close()
	db := session.DB(r_dbName)

	// angeforderte GridFs-collection dieser DB:
	gridfs := db.GridFS(r_gridfsName)

	// file aus GridFS lesen und als response senden:
	gridFile, err := gridfs.Open(r_fileName)
	check_ResponseToHTTP(err, w)

	// content-type header senden:
	tmpSlice := strings.Split(r_fileName, ".")
	fileExtension := tmpSlice[len(tmpSlice)-1] // das letzte Element
	fileExtension = strings.ToLower(fileExtension)
	var mimeType string
	switch fileExtension {
	case "jpeg", "jpg":
		mimeType = "image/jpeg"
	case "png":
		mimeType = "image/png"
	case "gif":
		mimeType = "image/gif"
	default:
		mimeType = "text/html"
	}
	w.Header().Add("Content-Type", mimeType)

	// image senden:
	_, err = io.Copy(w, gridFile)
	check_ResponseToHTTP(err, w)

	err = gridFile.Close()
	check_ResponseToHTTP(err, w)

}

//Der aktuelle Pool wird als Templet angezeigt
func poolAnzeige(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	poolname := r.FormValue("sammlungname")

	keks, _ := r.Cookie("user")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	for i := 0; i < len(ergDatenLog.Pools); i++ {
		if poolname == ergDatenLog.Pools[i].Name {
			t.ExecuteTemplate(w, "bilderAnzeige.html", ergDatenLog.Pools[i])
		}
	}
}

//Die aktuelle Sammlung wird als Templet angezeigt
func sammlungAnzeige(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	sammlungname := r.FormValue("sammlungname")

	keks, _ := r.Cookie("user")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	for i := 0; i < len(ergDatenLog.Sammlung); i++ {
		if sammlungname == ergDatenLog.Sammlung[i].Name {
			t.ExecuteTemplate(w, "bilderAnzeigeVonSammlung.html", ergDatenLog.Sammlung[i])
		}
	}
}

//DIe pool auswahl wird hier
func poolAuswahl(w http.ResponseWriter, r *http.Request) {
	//brauche allerdings eine Zufalls zahl da sonst das Histogramm bild im Cache geladen wird
	type poolTypTemp struct {
		Name          string
		GenBilder     bool
		Eigenschaften poolEigenschaftenTyp
		Bilder        []bilderTyp
		Zufall        int
	}
	type poolsTypTemp struct {
		Name           string
		Pools          []poolTypTemp
		Sammlung       []sammlungTyp
		MosaikSammlung mosaikTyp
	}

	keks, _ := r.Cookie("user")

	var ergDatenLog poolsTyp
	var ergDatenLog2 poolsTypTemp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	ergDatenLog2.Name = ergDatenLog.Name
	ergDatenLog2.MosaikSammlung = ergDatenLog.MosaikSammlung
	ergDatenLog2.Sammlung = ergDatenLog.Sammlung
	//Hier wird alles in das tempArray kopiert welches dann nachher ins Templet mit gebene wird
	for i := 0; i < len(ergDatenLog.Pools); i++ {
		var zufall = int(rand.Float32() * 10000)
		var tempArray poolTypTemp
		tempArray.Bilder = ergDatenLog.Pools[i].Bilder
		tempArray.GenBilder = ergDatenLog.Pools[i].GenBilder
		tempArray.Eigenschaften = ergDatenLog.Pools[i].Eigenschaften
		tempArray.Name = ergDatenLog.Pools[i].Name
		tempArray.Zufall = zufall
		ergDatenLog2.Pools = append(ergDatenLog2.Pools, tempArray)
	}

	t.ExecuteTemplate(w, "poolButtonErstellen.html", ergDatenLog2)
}

//anzeigen aller sammlungen als button
func sammlungAuswahl(w http.ResponseWriter, r *http.Request) {

	keks, _ := r.Cookie("user")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	t.ExecuteTemplate(w, "sammlungButtonErstellen.html", ergDatenLog)
}

//anzeige der sammlungen bei der msaik erstellung
func sammlungAuswahlBeiMosaikAnzeige(w http.ResponseWriter, r *http.Request) {
	keks, _ := r.Cookie("user")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	//muss ein anderes Templet sein da hier kein delete button dabei ist
	t.ExecuteTemplate(w, "sammlungButtonErstellenMosaikAnzeige.html", ergDatenLog)
}

//wird ein neuer Pool erstellt
func neuerPoolErstellen(w http.ResponseWriter, r *http.Request) {
	os.Mkdir("./temps", 0700)
	keks, _ := r.Cookie("user")
	r.ParseForm()
	user := keks.Value
	poolname := r.FormValue("name")
	bildergroesse, _ := strconv.Atoi(r.FormValue("bildergroesse"))
	bildHisto := "histoBild_" + poolname + user
	var genBilder bool
	if r.FormValue("pooltypeauswahl") == "uploadpooltypeauswahl" {
		genBilder = false
	} else {
		genBilder = true
	}
	//sucht den eintrag vom akktuellen user
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	//pools einzigartig machen durch abfrage ob dieser schon exestiert
	for i := 0; i < len(ergDatenLog.Pools); i++ {
		if poolname == ergDatenLog.Pools[i].Name {
			fmt.Fprint(w, "doppelterEintrag")
			return
		}
	}

	var tmpEigenschaften poolEigenschaftenTyp
	var tmpBild []bilderTyp
	//eigene bidler Pool
	if genBilder == false {

		tmpEigenschaften = poolEigenschaftenTyp{Anzahl: 0, KachelGroesse: bildergroesse, Helligkeit: 0, Farbverlauf: false, HistoBild: bildHisto}

		//neuen Pool hinzufügen
		ergDatenLog.Pools = append(ergDatenLog.Pools, poolTyp{Name: poolname, GenBilder: genBilder, Eigenschaften: tmpEigenschaften, Bilder: tmpBild})
		err = pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, ergDatenLog)
	} else {
		//genierter Pool mti seinen Eigenschaften
		anzahl, _ := strconv.Atoi(r.FormValue("anzahl"))
		var farbverlauf bool
		if r.FormValue("farbig") == "farbverlauf" {
			farbverlauf = true
		} else {
			farbverlauf = false
		}
		helligkeit, _ := strconv.Atoi(r.FormValue("helligkeit"))
		tmpEigenschaften = poolEigenschaftenTyp{Anzahl: anzahl, KachelGroesse: bildergroesse, Helligkeit: helligkeit, Farbverlauf: farbverlauf, HistoBild: bildHisto}

		//neuen Pool hinzufügen
		ergDatenLog.Pools = append(ergDatenLog.Pools, poolTyp{Name: poolname, GenBilder: genBilder, Eigenschaften: tmpEigenschaften, Bilder: tmpBild})
		err = pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, ergDatenLog)
		//Hier wird dann die funktion Kacheln genieren angezeigt
		kachelnGenerieren(poolname, bildergroesse, anzahl, farbverlauf, helligkeit, keks.Value)
	}
	prozentAnzeigeZahl = 99
	//dann wird noch ein histogramm Bild erzeugt da ja auch schon der Pool mit generierten Kachelen gefüllt sein kann
	macheHistoBildPool(poolname, user)
	prozentAnzeigeZahl = 0
	os.RemoveAll("./temps/")
}

//Hier werden die einstellungen für das mosika angezeigt in einem Templete bzw die datein mit rein gegeben
func mosaikEinstellungGenBilder(w http.ResponseWriter, r *http.Request) {

	keks, _ := r.Cookie("user")
	sammlungName, _ := r.Cookie("aktiveSammlung")
	r.ParseForm()
	bildNameDB := r.FormValue("bildname")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	for i := 0; i < len(ergDatenLog.Sammlung); i++ {
		if sammlungName.Value == ergDatenLog.Sammlung[i].Name {
			for j := 0; j < len(ergDatenLog.Sammlung[i].Bilder); j++ {
				if ergDatenLog.Sammlung[i].Bilder[j].BildNameDB == bildNameDB {

					t.ExecuteTemplate(w, "mosaikEinstellungGenBilderAnzeige.html", ergDatenLog.Sammlung[i].Bilder[j])
				}
			}
		}
	}
}

// Hier werden alle pools des aktuellen users gesucht und in ein Templet geladen
func gibAllePoolsMosaik(w http.ResponseWriter, r *http.Request) {
	keks, _ := r.Cookie("user")

	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	t.ExecuteTemplate(w, "allePoolsAnzeigen.html", ergDatenLog)

}

//ALle mosaike werden angezeigt
func mosaikShowSammlung(w http.ResponseWriter, r *http.Request) {

	keks, _ := r.Cookie("user")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	t.ExecuteTemplate(w, "bilderAnzeigeVonMosaikSammlung.html", ergDatenLog.MosaikSammlung)

}

//Hier werden die Infromationen zu dem Mosaik bidl rausgesucht
func infoMosaikBild(w http.ResponseWriter, r *http.Request) {
	keks, _ := r.Cookie("user")
	r.ParseForm()
	bildNameDB := r.FormValue("nameDB")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	for i := 0; i < len(ergDatenLog.MosaikSammlung.Bilder); i++ {
		if bildNameDB == ergDatenLog.MosaikSammlung.Bilder[i].MosaikNameDB {
			t.ExecuteTemplate(w, "mosaikBildGros.html", ergDatenLog.MosaikSammlung.Bilder[i])
		}
	}

}

//Hier wird das eigentliche mosaik erstellt
func machMosaik(w http.ResponseWriter, r *http.Request) {
	//temps Ordner erstellen
	os.Mkdir("./temps", 0700)
	keks, _ := r.Cookie("user")
	aktiveSammlungKeks, _ := r.Cookie("aktiveSammlung")
	// GridFs-collection "bilder" dieser DB:
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)
	r.ParseForm()
	orgiBildName := r.FormValue("mosaikVorlageName")
	mosaikVorlageName := r.FormValue("mosaikVorlageNameDB")
	var kachelVerwendungMehrmals bool
	kachelVerwendungMehrmalsString := r.FormValue("kachelVerwendungMehrmals")
	if kachelVerwendungMehrmalsString == "true" {
		kachelVerwendungMehrmals = true
	} else {
		kachelVerwendungMehrmals = false
	}
	//umwandeln von string in Int wert
	nBesteKacheln, _ := strconv.Atoi(r.FormValue("nBesteKacheln"))

	poolName := r.FormValue("poolName")
	//wird noch gesetzt
	kachelGroesse := 0
	//Variablen für Skalierung
	//70 / 80 wäre optimal bzw gut
	basisBildResizeX := 30
	basisBildResizeY := 40

	gibGridFile, _ := gridfs.Open(mosaikVorlageName)
	//neues File im Temp Oderner erzeugen
	tempBildRoh, _ := os.Create("./temps/tempBildRoh.jpg")

	_, err = io.Copy(tempBildRoh, gibGridFile)

	bildRoh, _ := imaging.Open("./temps/tempBildRoh.jpg")
	alteBreite := bildRoh.Bounds().Size().X
	alteHoehe := bildRoh.Bounds().Size().Y
	bildRoh = imaging.Resize(bildRoh, basisBildResizeX, basisBildResizeY, imaging.Lanczos)

	_ = imaging.Save(bildRoh, "./temps/tempBildRoh.jpg")

	bildRoh, _ = imaging.Open("./temps/tempBildRoh.jpg")

	//****************Pool Bekommen******************
	var poolBilder []bilderTyp
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)

	var gewaehlterPool poolTyp

	for i := 0; i < len(ergDatenLog.Pools); i++ {
		if poolName == ergDatenLog.Pools[i].Name {
			gewaehlterPool = ergDatenLog.Pools[i]
			kachelGroesse = ergDatenLog.Pools[i].Eigenschaften.KachelGroesse
			for j := 0; j < len(ergDatenLog.Pools[i].Bilder); j++ {
				poolBilder = append(poolBilder, ergDatenLog.Pools[i].Bilder[j])
			}
		}
	}
	//*********************************************

	//Prüfe ob einmalige Verwendung geht
	//Zu wenig Kacheln im Pool
	if kachelVerwendungMehrmals == false {
		needKacheln := (basisBildResizeX * basisBildResizeY) + nBesteKacheln
		if needKacheln > len(poolBilder) {
			fmt.Fprint(w, "zuWenigKacheln")
			return
		}
	}
	if len(poolBilder) < nBesteKacheln {
		fmt.Fprint(w, "zuWenigKacheln")
		return
	}

	var passendeKachel []passendeKachelTyp
	//Hier wird die Breite des Fertigen Mosaik bildes berechnet
	mosaikBildBreite := basisBildResizeX * kachelGroesse
	mosaikBildHoehe := basisBildResizeY * kachelGroesse
	//erste ist minX , minY maxX, max Y
	startMalen := image.Rect(0, 0, kachelGroesse, kachelGroesse)
	// Mosaik Vorlage als leeres Bild
	newMosaik := image.NewRGBA(image.Rect(0, 0, mosaikBildBreite, mosaikBildHoehe))
	_ = imaging.Save(newMosaik, "./temps/mosaikBild.jpg")

	//wird für die lade anzeige gebraucht
	count := 0
	anzahlProzent := bildRoh.Bounds().Size().Y * bildRoh.Bounds().Size().X

	for x := 0; x < bildRoh.Bounds().Size().Y; x++ {
		for y := 0; y < bildRoh.Bounds().Size().X; y++ {
			// RGB Werte jedes Pixels des RohBildes
			rTemp, gTemp, bTemp, _ := bildRoh.At(y, x).RGBA()
			//Geht jedes Poolbild durch
			//Mittlerer Farbwert der Poolbilder - Farbe eines Basismotivpixels
			for i := 0; i < len(poolBilder); i++ {
				//Farbe muss geshiftet werden damit er es als int erkennt
				//anderenfalls ergibt 95-108 4milliarden
				tempR := poolBilder[i].MittlererFarbWert[0] >> 8
				tempR2 := rTemp >> 8
				dR := int(tempR) - int(tempR2)
				//muss von unsigend weg damit man als in in den minus bereich gehen kann
				tempG := poolBilder[i].MittlererFarbWert[1] >> 8
				tempG2 := gTemp >> 8
				dG := int(tempG) - int(tempG2)

				tempB := poolBilder[i].MittlererFarbWert[2] >> 8
				tempB2 := bTemp >> 8
				dB := int(tempB) - int(tempB2)

				//Vektorlänge
				d := math.Sqrt(sq(float64(dR)) + sq(float64(dG)) + sq(float64(dB)))
				//Namen der Kachel und Länge des Vektors
				passendeKachel = append(passendeKachel, passendeKachelTyp{KleinsteD: d, Name: poolBilder[i].Bild})

			}
			//sortieren der passenden Kacheln nach der Länge des Vektors kleinester d Wert ist dann oben
			sort.Sort(SortByKleinstesD(passendeKachel))
			//hier wird der lade balken zahl neu berechnet
			count++
			prozentAnzeigeZahl = (count * 100) / anzahlProzent
			//n Besten Kacheln wählen, falls ausgewählt, sonst das kleinste d genommen
			var tempKachelNummer int

			if nBesteKacheln != 1 {
				//zufalls zahl zwischen dem besten und dem n Wert gebildet um die Kachel zu ermitteln
				tempKachelNummer = random(int(0), nBesteKacheln)
			} else {
				tempKachelNummer = 0
			}
			//Öffnen der passenden Kachel
			//Kachelnummer kommt aus sortierter Liste
			//an der n Besten Stelle
			gibGridFileKachel, err := gridfs.Open(passendeKachel[tempKachelNummer].Name)
			check(err)
			defer gibGridFileKachel.Close()
			//mehrmalsVerwenden der Kacheln
			//Ausorteiren der grade eben Verwendeten Kachel aus dem poolBilder Array
			if kachelVerwendungMehrmals == false {
				var tempArray []bilderTyp
				for i := 0; i < len(poolBilder); i++ {
					if passendeKachel[tempKachelNummer].Name != poolBilder[i].Bild {
						tempArray = append(tempArray, poolBilder[i])
					}
				}
				poolBilder = tempArray
			}
			//hier wird das array wieder geleert
			passendeKachel = nil

			tempKachelRoh, _ := os.Create("./temps/tempKachel.jpg")
			_, err = io.Copy(tempKachelRoh, gibGridFileKachel)
			tempKachel, _ := imaging.Open("./temps/tempKachel.jpg")
			//Einfügen der passenden Kachel
			draw.Draw(newMosaik, startMalen, tempKachel, image.Point{0, 0}, draw.Src)
			//erhöht den Breite um Kachelgröße
			startMalen.Min.X = startMalen.Min.X + kachelGroesse
			startMalen.Max.X = startMalen.Max.X + kachelGroesse
			tempKachelRoh.Close()
		}
		//Setze Wert zurück damit die Breite wieder komplett gemalt wird
		startMalen.Min.X = 0
		startMalen.Max.X = kachelGroesse
		//und zählt einen Runter damit die vorher gemalte rheie nicht überschrieben wird
		startMalen.Min.Y = startMalen.Min.Y + kachelGroesse
		startMalen.Max.Y = startMalen.Max.Y + kachelGroesse

	}
	//Speicherung der Bilder
	_ = imaging.Save(newMosaik, "./temps/mosaikBild.jpg")
	thumbMosaik, _ := imaging.Open("./temps/mosaikBild.jpg")
	imgThumbMosaik := imaging.Thumbnail(thumbMosaik, 200, 200, imaging.CatmullRom)
	imaging.Save(imgThumbMosaik, "./temps/mosaikBildThumb.jpg")

	mosaikHoheTemp := thumbMosaik.Bounds().Size().Y
	mosaikBreiteTemp := thumbMosaik.Bounds().Size().X

	//********************************datenBank************************************
	var mosaikBild mosaikBilderTyp
	var tempBasisBild sammlungBilderTyp
	tempBasisBild = sammlungBilderTyp{BildName: orgiBildName, BildNameDB: mosaikVorlageName, Hohe: alteHoehe, Breite: alteBreite}

	mosaikName := "mosaik_" + orgiBildName
	mosaikNameDB := "mosaik_" + mosaikVorlageName
	mosaikNameThumb := "mosaik_Thumb" + mosaikNameDB

	//*****************Namensfindung*****************
	erg := make(map[string]interface{})
	iter := gridfs.Find(nil).Iter()
	zahl := 1
	for iter.Next(erg) {

		for erg["filename"] == mosaikNameDB {
			mosaikNameDB = strconv.Itoa(zahl) + mosaikNameDB
			mosaikNameThumb = strconv.Itoa(zahl) + mosaikNameThumb
			zahl++
		}
	}
	// in datenbank schreiben
	mosaikBild = mosaikBilderTyp{MosaikName: mosaikName, MosaikNameDB: mosaikNameDB, MosaikHoehe: mosaikHoheTemp, MosaikBreite: mosaikBreiteTemp, BasisBild: tempBasisBild, PoolEigenschaften: gewaehlterPool, DoppelteVerwendung: kachelVerwendungMehrmals, NbestGeeignet: nBesteKacheln, ThumbBild: mosaikNameThumb, SammlungName: aktiveSammlungKeks.Value}
	ergDatenLog.MosaikSammlung.Bilder = append(ergDatenLog.MosaikSammlung.Bilder, mosaikBild)
	pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, ergDatenLog)

	// grid-file mit diesem Namen erzeugen:
	gridFile, _ := gridfs.Create(mosaikNameDB)
	gridFileThumb, _ := gridfs.Create(mosaikNameThumb)
	bildMosaik, _ := os.Open("./temps/mosaikBild.jpg")
	bildMosaikThumb, _ := os.Open("./temps/mosaikBildThumb.jpg")

	// in GridFSkopieren:
	io.Copy(gridFile, bildMosaik)
	io.Copy(gridFileThumb, bildMosaikThumb)

	gridFile.Close()
	gridFileThumb.Close()

	gibGridFile.Close()
	prozentAnzeigeZahl = 0

	tempBildRoh.Close()
	bildMosaikThumb.Close()
	bildMosaik.Close()
	os.RemoveAll("./temps/")

}

//zufalls zahl mit interger werten und begrenzungen nach min und max
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

//Das ist eine funktion für quadrat wurzeln
func sq(v float64) float64 {
	return v * v
}

//hier wird ein absis bild eine neue größe gegeben
func resizeBild(w http.ResponseWriter, r *http.Request) {
	os.Mkdir("./temps", 0700)
	var bilderGrose int
	bilderGrose = 150

	// GridFs-collection "bilder" dieser DB:
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)

	r.ParseForm()
	bildNameDB := r.FormValue("bildname")
	breiteString := r.FormValue("breite")
	hoeheString := r.FormValue("hoehe")

	sammlungName := r.FormValue("sammlungname")
	fileName := breiteString + "x" + hoeheString + "_" + bildNameDB
	breite, _ := strconv.Atoi(breiteString)
	hoehe, _ := strconv.Atoi(hoeheString)

	//*****************Namensfindung*****************
	originalBildName := fileName
	erg := make(map[string]interface{})
	iter := gridfs.Find(nil).Iter()
	zahl := 1
	for iter.Next(erg) {

		for erg["filename"] == fileName {
			fileName = strconv.Itoa(zahl) + fileName
			zahl++
		}
	}
	//********************Ende***********************

	// file aus GridFS lesen und als response senden:
	gibGridFile, _ := gridfs.Open(bildNameDB)

	//neues File im Temp Oderner erzeugen
	tempBildRoh, _ := os.Create("./temps/tempBildRoh.jpg")

	_, err = io.Copy(tempBildRoh, gibGridFile)

	//Bild aus dem Temp öffnen
	imgRoh, err := os.Open("./temps/tempBildRoh.jpg")
	check_ResponseToHTTP(err, w)

	imgTemp, _, _ := image.Decode(imgRoh)

	img := imaging.Resize(imgTemp, breite, hoehe, imaging.Box)

	_ = imaging.Save(img, "./temps/tempBildQuadrat.jpg")

	//*************HistoBild******************

	histoFileName := "histo_" + fileName
	machHistoBildSammlung()

	bildHisto, _ := os.Open("./temps/tempBildHisto.jpg")

	// grid-file mit diesem Namen erzeugen:
	gridFileHisto, err := gridfs.Create(histoFileName)
	check_ResponseToHTTP(err, w)

	// in GridFSkopieren:
	_, _ = io.Copy(gridFileHisto, bildHisto)
	_ = gridFileHisto.Close()
	//****************************************

	imgFertig, err := os.Open("./temps/tempBildQuadrat.jpg")
	// grid-file mit diesem Namen erzeugen:
	gridFile, err := gridfs.Create(fileName)
	check_ResponseToHTTP(err, w)

	// in GridFSkopieren:
	_, err = io.Copy(gridFile, imgFertig)
	check_ResponseToHTTP(err, w)

	err = gridFile.Close()
	check_ResponseToHTTP(err, w)

	//************thumb machen*****************

	tempBildRoh, _ = os.Open("./temps/tempBildQuadrat.jpg")
	imgTempThumb, _, _ := image.Decode(tempBildRoh)

	img = imaging.Thumbnail(imgTempThumb, bilderGrose, bilderGrose, imaging.CatmullRom)
	err = imaging.Save(img, "./temps/tempBildFertig.jpg")

	imgFertigThumb, err := os.Open("./temps/tempBildFertig.jpg")
	// grid-file mit diesem Namen erzeugen:
	gridFileThumb, err := gridfs.Create("thumbBild" + fileName)
	check_ResponseToHTTP(err, w)

	// in GridFSkopieren:
	_, err = io.Copy(gridFileThumb, imgFertigThumb)
	check_ResponseToHTTP(err, w)

	err = gridFileThumb.Close()
	check_ResponseToHTTP(err, w)

	//************db schreiben ************************

	//***********************Anker damit Bilder den User/Sammlungen zugeordnet werden können *****************
	keks, _ := r.Cookie("user")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog) // entspricht findOne

	//hochgeladene Bild auch in der COllection mit Name versehen
	for j := 0; j < len(ergDatenLog.Sammlung); j++ {

		if sammlungName == ergDatenLog.Sammlung[j].Name {
			ergDatenLog.Sammlung[j].Bilder = append(ergDatenLog.Sammlung[j].Bilder,
				sammlungBilderTyp{BildName: originalBildName, BildNameDB: fileName, Hohe: hoehe, Breite: breite, HistoBild: histoFileName, ThumbBild: "thumbBild" + fileName})
			pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, ergDatenLog)

		}
	}
	//*****************************************************************************************************

	imgFertigThumb.Close()
	imgFertig.Close()
	imgRoh.Close()
	bildHisto.Close()
	tempBildRoh.Close()
	os.RemoveAll("./temps/")
}

//hier werden infos der bilder aus der sammlung in ein templet geladen
func gibBildInfo(w http.ResponseWriter, r *http.Request) {
	keks, _ := r.Cookie("user")
	r.ParseForm()
	bildNameDB := r.FormValue("bildname")
	sammlungName := r.FormValue("sammlungname")
	//durchsuchen der DB
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	for i := 0; i < len(ergDatenLog.Sammlung); i++ {
		if sammlungName == ergDatenLog.Sammlung[i].Name {
			for j := 0; j < len(ergDatenLog.Sammlung[i].Bilder); j++ {
				if ergDatenLog.Sammlung[i].Bilder[j].BildNameDB == bildNameDB {

					t.ExecuteTemplate(w, "aktivesBildInfo.html", ergDatenLog.Sammlung[i].Bilder[j])
				}
			}
		}
	}

}

//zufällige kacheln genierieren
func kachelnGenerieren(poolname_ string, bildergroesse_ int, anzahl_ int, farbverlauf_ bool, helligkeit_ int, user_ string) {

	// GridFs-collection "bilder" dieser DB:
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)
	user := user_
	poolname := poolname_
	bildergroesse := bildergroesse_
	farbverlauf := farbverlauf_
	helligkeit := helligkeit_
	anzahl := anzahl_
	genBildName := "GenBild"
	//wird für dei Porzent anzeige gebraucht
	count := 0
	anzahlProzent := anzahl

	for i := 0; i < anzahl; i++ {
		//prozent zahl wird neu berechnet
		count++
		prozentAnzeigeZahl = (count * 100) / anzahlProzent
		var col color.RGBA
		//zufalls farben werden mit dem helligkeits faktor erstellt
		zufallRed := uint8(rand.Intn(255) * helligkeit / 127)
		zufallGreen := uint8(rand.Intn(255) * helligkeit / 127)
		zufallBlue := uint8(rand.Intn(255) * helligkeit / 127)

		col = color.RGBA{zufallRed, zufallGreen, zufallBlue, 0}

		// image generieren:
		breite := bildergroesse
		hoehe := bildergroesse
		img := image.NewRGBA(image.Rect(0, 0, breite, hoehe))
		//wenn farbverlauf
		if farbverlauf {
			for x := 0; x < breite; x++ {
				for y := 0; y < hoehe; y++ {
					//hier wird mit der zufalls Farbe der verlauf bestimmt
					schritteRed := (255 - zufallRed) / uint8(breite)
					schritteGreen := (255 - zufallGreen) / uint8(breite)
					schritteBlue := (255 - zufallBlue) / uint8(breite)
					//start wert ab wo der farb verlauf anfangen soll
					redCol := zufallRed + uint8(y*x/breite)*(schritteRed)
					greenCol := zufallGreen + uint8(y)*(schritteGreen)
					blueCol := zufallBlue + uint8(x)*(schritteBlue)
					//auf auf mache farbe
					col = color.RGBA{redCol, greenCol, blueCol, 0}
					img.Set(x, y, col)
				}
			}
		} else {
			// gesamtes image füllen: mit der farbe :)
			for x := 0; x < breite; x++ {
				for y := 0; y < hoehe; y++ {
					img.Set(x, y, col)
				}
			}
		}

		//neues File im Temp Ordner erzeugen
		tempBildQuadrat, _ := os.Create("./temps/tempBildQuadrat.jpg")
		//wird als jpeg gekennzeichnet
		jpeg.Encode(tempBildQuadrat, img, &jpeg.Options{jpeg.DefaultQuality})
		tempBildQuadrat.Close()
		//Bild aus dem Temp öffnen
		imgFertig, _ := os.Open("./temps/tempBildQuadrat.jpg")

		bildName := poolname + genBildName + strconv.Itoa(i)
		// grid-file mit diesem Namen erzeugen:
		gridFile, _ := gridfs.Create(bildName)
		// in GridFSkopieren:
		_, _ = io.Copy(gridFile, imgFertig)
		_ = gridFile.Close()

		//******************MittlererFarbWert*********************

		var r uint32
		var g uint32
		var b uint32
		r, g, b = 0, 0, 0
		mittleWertImg, _ := imaging.Open("./temps/tempBildQuadrat.jpg")
		for i := 0; i < mittleWertImg.Bounds().Size().Y; i++ {
			for j := 0; j < mittleWertImg.Bounds().Size().X; j++ {
				rTemp, gTemp, bTemp, _ := mittleWertImg.At(j, i).RGBA()
				r, g, b = r+uint32(rTemp), g+uint32(gTemp), b+uint32(bTemp)
			}
		}
		allePixels := uint32(mittleWertImg.Bounds().Max.X * mittleWertImg.Bounds().Max.Y)

		mittlereWert := [3]uint32{r / allePixels, g / allePixels, b / allePixels}
		//********************************************************

		//*************************anker*************
		var ergDatenLog poolsTyp
		pools.Find(bson.M{"name": bson.M{"$eq": user}}).One(&ergDatenLog) // entspricht findOne

		for j := 0; j < len(ergDatenLog.Pools); j++ {
			if poolname == ergDatenLog.Pools[j].Name {

				ergDatenLog.Pools[j].Bilder = append(ergDatenLog.Pools[j].Bilder, bilderTyp{Bild: bildName, MittlererFarbWert: mittlereWert})
				pools.Update(bson.M{"name": bson.M{"$eq": user}}, ergDatenLog)

			}
		}

		imgFertig.Close()
	}

}

//hier wird eine neu smamlung angelegt
func neueSammlungErstellen(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	sammlungname := r.FormValue("name")

	keks, _ := r.Cookie("user")
	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	//falls es die sammlung schon gibt
	for i := 0; i < len(ergDatenLog.Sammlung); i++ {
		if sammlungname == ergDatenLog.Sammlung[i].Name {
			fmt.Fprint(w, "doppelterEintrag")
		}
	}
	var tmpBild []sammlungBilderTyp
	//neue Sammlung hinzufügen
	ergDatenLog.Sammlung = append(ergDatenLog.Sammlung, sammlungTyp{Name: sammlungname, Bilder: tmpBild})
	err = pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, ergDatenLog)

}

//löschen von Pools
func poolDelete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	poolname := r.FormValue("poolname")
	keks, _ := r.Cookie("user")
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)

	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	var tempPools poolsTyp
	tempPools.MosaikSammlung = ergDatenLog.MosaikSammlung
	tempPools.Sammlung = ergDatenLog.Sammlung
	tempPools.Name = ergDatenLog.Name
	//prozent anzeige
	prozentAnzeigeZahl = 0
	count := 0
	for i := 0; i < len(ergDatenLog.Pools); i++ {
		if poolname != ergDatenLog.Pools[i].Name {
			//pools appenden und durch auslassen des zu löschenden pools er rausfällt
			tempPools.Pools = append(tempPools.Pools, ergDatenLog.Pools[i])

		} else {
			gridfs.Remove(ergDatenLog.Pools[i].Eigenschaften.HistoBild)

			anzahlFiles := len(ergDatenLog.Pools[i].Bilder)
			for j := 0; j < len(ergDatenLog.Pools[i].Bilder); j++ {
				gridfs.Remove(ergDatenLog.Pools[i].Bilder[j].Bild)
				//prozent anziege
				count++
				prozentAnzeigeZahl = (count * 100) / anzahlFiles

			}

		}
	}
	//falls keine pools mehr da sind erstelle einen neuen test Pool
	if len(tempPools.Pools) == 0 {
		var tmpBild []bilderTyp
		tempName := keks.Value
		tempHistoName := "histoBild_PoolTest" + tempName
		var tempEigenschaften poolEigenschaftenTyp
		tempEigenschaften = poolEigenschaftenTyp{Anzahl: 0, KachelGroesse: 25, Helligkeit: 0, Farbverlauf: false, HistoBild: tempHistoName}
		tempPools.Pools = append(tempPools.Pools, poolTyp{Name: "PoolTest", GenBilder: false, Eigenschaften: tempEigenschaften, Bilder: tmpBild})
		fmt.Fprint(w, "error")
		//histoBildPool
		os.Mkdir("./temps", 0700)
		gridfsName := "bilder"
		gridfs := db.GridFS(gridfsName)
		imgHisto := image.NewRGBA(image.Rect(0, 0, 10, 10))
		tempBildHisto, _ := os.Create("./temps/tempBildHisto.jpg")
		jpeg.Encode(tempBildHisto, imgHisto, &jpeg.Options{jpeg.DefaultQuality})
		fertigHisto, _ := os.Open("./temps/tempBildHisto.jpg")
		gridFile, _ := gridfs.Create(tempHistoName)
		io.Copy(gridFile, fertigHisto)

		gridFile.Close()
		fertigHisto.Close()
		os.RemoveAll("./temps/")
	}
	//daten bank erneuern
	pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, tempPools)

}

//sammlung löschen
func sammlungDelete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	sammlungname := r.FormValue("sammlungname")
	keks, _ := r.Cookie("user")
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)

	var ergDatenLog poolsTyp
	pools.Find(bson.M{"name": bson.M{"$eq": keks.Value}}).One(&ergDatenLog)
	var tempPools poolsTyp
	tempPools.Pools = ergDatenLog.Pools
	tempPools.MosaikSammlung = ergDatenLog.MosaikSammlung
	tempPools.Name = ergDatenLog.Name
	prozentAnzeigeZahl = 0
	count := 0
	//auch hier wird das aray wieder zummen gebaut und das zu löschende weggelassen
	for i := 0; i < len(ergDatenLog.Sammlung); i++ {
		if sammlungname != ergDatenLog.Sammlung[i].Name {

			tempPools.Sammlung = append(tempPools.Sammlung, ergDatenLog.Sammlung[i])

		} else {

			anzahlFiles := len(ergDatenLog.Sammlung[i].Bilder)
			for j := 0; j < len(ergDatenLog.Sammlung[i].Bilder); j++ {
				count++
				prozentAnzeigeZahl = (count * 100) / anzahlFiles
				gridfs.Remove(ergDatenLog.Sammlung[i].Bilder[j].BildNameDB)
				gridfs.Remove(ergDatenLog.Sammlung[i].Bilder[j].ThumbBild)
				gridfs.Remove(ergDatenLog.Sammlung[i].Bilder[j].HistoBild)
			}

		}
	}
	//falls keine sammlung mehr da sit eine test sammlung erstellen
	if len(tempPools.Sammlung) == 0 {

		var tempBildSammlung []sammlungBilderTyp
		tempPools.Sammlung = append(tempPools.Sammlung, sammlungTyp{Name: "SammlungTest", Bilder: tempBildSammlung})
		fmt.Fprint(w, "error")

	}
	pools.Update(bson.M{"name": bson.M{"$eq": keks.Value}}, tempPools)

}

//registration
func macheRegestrierung(w http.ResponseWriter, r *http.Request) {
	check := func(err error) {
		if err != nil {
			fmt.Println(err)
		}
	}

	switch r.Method {
	case "GET":
		t.ExecuteTemplate(w, "registrieren", nil)

	case "POST":
		r.ParseForm()
		name := r.FormValue("name")
		pw1 := r.FormValue("pw1")
		pw2 := r.FormValue("pw2")

		var ergDatenLog userTyp
		var err = userPW.Find(bson.M{"name": bson.M{"$eq": name}}).One(&ergDatenLog) // entspricht findOne

		byteArray := []byte(name)
		byteArray2 := []byte(pw1)
		if len(byteArray) < 2 {
			t.ExecuteTemplate(w, "registrieren", "Name zu kurz")
			return
		}
		if len(byteArray2) < 4 {
			t.ExecuteTemplate(w, "registrieren", "Passwort zu kurz")
			return
		}
		for i := 0; i < len(byteArray); i++ {
			if (122 >= byteArray[i] && byteArray[i] >= 97) || (90 >= byteArray[i] && byteArray[i] >= 65) || (57 >= byteArray[i] && byteArray[i] >= 48) {

			} else {
				t.ExecuteTemplate(w, "registrieren", "Ungültiges Zeichen")
				return
			}
		}
		for i := 0; i < len(byteArray2); i++ {
			if (122 >= byteArray2[i] && byteArray2[i] >= 97) || (90 >= byteArray2[i] && byteArray2[i] >= 65) || (57 >= byteArray2[i] && byteArray2[i] >= 48) {

			} else {
				t.ExecuteTemplate(w, "registrieren", "Ungültiges Zeichen")
				return
			}
		}

		if pw1 == "" || pw2 == "" {
			t.ExecuteTemplate(w, "registrieren", "Passwort darf nicht leer sein")
		} else if name == "" {
			t.ExecuteTemplate(w, "registrieren", "Name darf nicht leer sein")
		} else if pw1 != pw2 {
			t.ExecuteTemplate(w, "registrieren", "Passwort stimmt nicht überein")
		} else if ergDatenLog.Name == name {
			t.ExecuteTemplate(w, "registrieren", "User bereits vorhanden")
		} else {

			var temp = userTyp{Name: name, Passwort: pw1}
			err = userPW.Insert(temp)
			check(err)

			var tmpBild []bilderTyp
			var tempBildSammlung []sammlungBilderTyp
			var tempSammlung []sammlungTyp
			var tmpPool []poolTyp
			var tempEigenschaften poolEigenschaftenTyp
			var tmpMosaikSammlung mosaikTyp
			//datenbanke einträge
			tempHistoName := "histoBild_PoolTest" + name
			//es wird ein test Pool und eine Test SAmmlung erstellt
			tempEigenschaften = poolEigenschaftenTyp{Anzahl: 0, KachelGroesse: 25, Helligkeit: 0, Farbverlauf: false, HistoBild: tempHistoName}
			tmpPool = append(tmpPool, poolTyp{Name: "PoolTest", GenBilder: false, Eigenschaften: tempEigenschaften, Bilder: tmpBild})
			tempSammlung = append(tempSammlung, sammlungTyp{Name: "SammlungTest", Bilder: tempBildSammlung})
			var temp2 = poolsTyp{Name: name, Pools: tmpPool, Sammlung: tempSammlung, MosaikSammlung: tmpMosaikSammlung}

			err = pools.Insert(temp2)
			check(err)

			t.ExecuteTemplate(w, "registrieren", "alles richtig")
			//histoBildPool
			os.Mkdir("./temps", 0700)
			gridfsName := "bilder"
			gridfs := db.GridFS(gridfsName)
			//Hier wird ein schwwarzes Histogramm bild erzeugt für einen leeren POol
			//damit das layout auch ein bild bekommt was es braucht
			imgHisto := image.NewRGBA(image.Rect(0, 0, 10, 10))
			tempBildHisto, _ := os.Create("./temps/tempBildHisto.jpg")
			jpeg.Encode(tempBildHisto, imgHisto, &jpeg.Options{jpeg.DefaultQuality})
			fertigHisto, _ := os.Open("./temps/tempBildHisto.jpg")
			gridFile, err := gridfs.Create(tempHistoName)
			io.Copy(gridFile, fertigHisto)

			fmt.Println(err)

			gridFile.Close()
			fertigHisto.Close()
			os.RemoveAll("./temps/")

		}
	}
}

//hier wird  die mosaik Anzeige geladen
func mosaikTab(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "showMosaikUI.html", nil)
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

//hier wird  die BasisiBild Anzeige geladen
func basisBildTab(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "showBasisUI.html", nil)
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

//hier wird  die Pool Anzeige geladen
func poolTab(w http.ResponseWriter, r *http.Request) {

	t.ExecuteTemplate(w, "showPoolUI.html", nil)
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

//Zum downloaden des mosaik bildes
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	os.Mkdir("./temps", 0700)
	bildNameDB := r.FormValue("fileName")
	bildName := r.FormValue("name")
	gridfsName := "bilder"
	gridfs := db.GridFS(gridfsName)
	gibGridFile, _ := gridfs.Open(bildNameDB)
	defer gibGridFile.Close()
	//neues File im Temp Oderner erzeugen
	tempBildRoh, _ := os.Create("./temps/tempBildFertig.jpg")

	_, err = io.Copy(tempBildRoh, gibGridFile)

	file, err := os.Open("./temps/tempBildFertig.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// fileName aus dateiUrl extrahieren:
	dateiName := ""
	if strings.Contains(bildName, "/") {
		split1 := strings.Split(bildName, "/")
		dateiName = split1[len(split1)-1]
	} else {
		dateiName = bildName
	}

	// Content-Type anhand der Dateiendung bestimmen:
	dateiExt := ""
	contentType := ""
	if strings.Contains(dateiName, ".") {
		split2 := strings.Split(dateiName, ".")
		dateiExt = split2[len(split2)-1]
	} else {
		dateiExt = "unbekannt"
	}

	dateiExt = strings.ToLower(dateiExt)
	switch dateiExt {
	case "txt":
		contentType = "text/plain"
	case "html":
		contentType = "text/html"
	case "gif":
		contentType = "image/gif"
	case "jpg":
		contentType = "image/jpeg"
	case "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	default:
		contentType = "application/octet-stream"

	}

	// file-size für Content-Length header bestimmen:
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
	}
	contentLength := fileInfo.Size()

	// Mit dem Content-Disposition header wird dem Browser mitgeteilt, die
	// folgende Datei nicht anzuzeigen, sondern in den download-Ordner zu kopieren:
	w.Header().Set("Content-Disposition", "attachment; filename="+dateiName)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))

	// file in den ResponseWriter kopieren:
	io.Copy(w, file)
	tempBildRoh.Close()
	os.RemoveAll("./temps/")
}

// hier wird die prozent zahl geholt und in ein template mit geben
func holProzent(w http.ResponseWriter, r *http.Request) {

	rand.Seed(int64(time.Now().Nanosecond()))
	var zufall = int(rand.Float32() * 10000)

	prozentPaket := prozentStatusTyp{
		//durch die zufalls zahl wird das bild nicht vom chache geladen
		Zufall:  zufall,
		Prozent: prozentAnzeigeZahl,
	}
	t.ExecuteTemplate(w, "prozentAnzeige.html", prozentPaket)
}

//das eignetlihc bild für die alde anzeige wird hier gemalt und zurück gebeben
func showLadeAnzeige(w http.ResponseWriter, r *http.Request) {

	getProzent := r.URL.Query().Get("prozent")
	gProzent, _ := strconv.Atoi(getProzent)

	img := image.NewRGBA(image.Rect(0, 0, 30, 200))
	// ein balken von unten nach oben der immer größer wird
	for x := 0; x < 30; x++ {
		for y := 200; y >= 200-(gProzent*2); y-- {
			img.Set(x, y, color.RGBA{66, 232, 35, 255})
		}
	}

	w.Header().Set("Content-type", "image/png")
	png.Encode(w, img)

}

func main() {

	http.HandleFunc("/gridGetImage", getImageHandler)
	http.HandleFunc("/gridMultiForm", multipartFormHandler)
	http.HandleFunc("/uploadBilderSammlung", uploadBilderSammlung)
	http.HandleFunc("/showLadeAnzeige", showLadeAnzeige)
	http.HandleFunc("/holProzent", holProzent)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/gibBildInfo", gibBildInfo)
	http.HandleFunc("/resizeBild", resizeBild)
	http.HandleFunc("/poolDelete", poolDelete)
	http.HandleFunc("/sammlungDelete", sammlungDelete)
	http.HandleFunc("/poolAnzeige", poolAnzeige)
	http.HandleFunc("/poolAuswahl", poolAuswahl)
	http.HandleFunc("/neuerPoolErstellen", neuerPoolErstellen)
	http.HandleFunc("/sammlungAnzeige", sammlungAnzeige)
	http.HandleFunc("/sammlungAuswahl", sammlungAuswahl)
	http.HandleFunc("/sammlungAuswahlBeiMosaikAnzeige", sammlungAuswahlBeiMosaikAnzeige)
	http.HandleFunc("/neueSammlungErstellen", neueSammlungErstellen)
	http.HandleFunc("/mosaikTab", mosaikTab)
	http.HandleFunc("/basisBildTab", basisBildTab)
	http.HandleFunc("/poolTab", poolTab)
	http.HandleFunc("/mosaikEinstellungGenBilder", mosaikEinstellungGenBilder)
	http.HandleFunc("/gibAllePoolsMosaik", gibAllePoolsMosaik)
	http.HandleFunc("/mosaikShowSammlung", mosaikShowSammlung)
	http.HandleFunc("/machMosaik", machMosaik)
	http.HandleFunc("/infoMosaikBild", infoMosaikBild)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/index", loginSeite)
	http.HandleFunc("/tessera", loginSeite)
	http.HandleFunc("/login", loginSeite)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/registrieren", macheRegestrierung)
	http.Handle("/", http.FileServer(http.Dir("./daten")))
	http.ListenAndServe(":4242", nil)
	defer dbSession.Close()
}
