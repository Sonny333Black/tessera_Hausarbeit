function sammlungJS() {
	//var serverPfad = "http://localhost:4242";
	var serverPfad="";
	//händeler regstieren und sammlung aktiv setzten
    var aktuellerSammlungName = "SammlungTest";
    var sammlungPfad = serverPfad+"/sammlungAnzeige?sammlungname=";
    var sammlung = new XMLHttpRequest(); //wird zum sammlung anzeigne gebraucht
    var neueSammlungErstellen = new XMLHttpRequest();//wird zum erstellen einer neuen sammlung gebraucht
    var sammlungAuswahl = new XMLHttpRequest(); // die auswahl zwishcen den sammlung wird hier abgehandelt
    var bildInfo = new XMLHttpRequest(); // die bild infomationen werden hier behandelt
    var resize = new XMLHttpRequest(); // resizes des bildes
    var sammlungDelete = new XMLHttpRequest(); // zum lschen einer sammlung
    var uploadSammlung = new XMLHttpRequest(); // hochladen von Bildern
    var ladeAnzeige = new XMLHttpRequest(); // lade anzeige
    var holProzent = new XMLHttpRequest(); // przent zahl der ladeanzeige bekommen
    var intervalHandle;
    sammlung.onreadystatechange = function () {
        if (sammlung.readyState === 4 && sammlung.status === 200) {
            document.getElementById("sammlungAnzeige").innerHTML = sammlung.responseText;
            addClickListenerBild();
        }
    }
    ladeAnzeige.onreadystatechange = function () {
        if (ladeAnzeige.readyState === 4 && ladeAnzeige.status === 200) {
            document.getElementById("ladeAnzeige").innerHTML = ladeAnzeige.responseText;

        }
    }
	// sammlung Auswahl anzeigen
    sammlungAuswahl.onreadystatechange = function () {
        if (sammlungAuswahl.readyState === 4 && sammlungAuswahl.status === 200) {
            document.getElementById("bilderSammlungAuswahl").innerHTML = sammlungAuswahl.responseText;
            addClickListener();
        }
    }
    
	uploadSammlung.onreadystatechange = function () {
        if (uploadSammlung.readyState === 4 && uploadSammlung.status === 200) {

            endLadeBalken();
            sammlungAnzeigen();


        }
    }
    sammlungDelete.onreadystatechange = function () {
        if (sammlungDelete.readyState === 4 && sammlungDelete.status === 200) {
            if (sammlungDelete.responseText == "error") {
                alert("Sie haben keine Sammlung mehr, es wird eine TestSammlung erstellt.");
            }
            bildInfo.open("GET", serverPfad+"/gibBildInfo?bildname=nil&sammlungname=nil");
            bildInfo.send();
            sammlungAuswahl.open("GET", serverPfad+"/sammlungAuswahl");
            sammlungAuswahl.send();
            sammlungAnzeigen();

        }
        endLadeBalken();
    }
    bildInfo.onreadystatechange = function () {
        if (bildInfo.readyState === 4 && bildInfo.status === 200) {
            document.getElementById("aktivesBild").innerHTML = bildInfo.responseText;
            setzteKlickUmZuSkalieren();
        }
    }
    neueSammlungErstellen.onreadystatechange = function () {
        if (neueSammlungErstellen.readyState === 4 && neueSammlungErstellen.status === 200) {
            if ("doppelterEintrag" == neueSammlungErstellen.responseText) {
                alert("Sammlung exestiert bereits");

                return;
            }

            sammlungAuswahl.open("GET", serverPfad+"/sammlungAuswahl");
            sammlungAuswahl.send();
            sammlungAnzeigen();

        }
    }
    resize.onreadystatechange = function () {
        if (resize.readyState === 4 && resize.status === 200) {
            sammlungAnzeigen();
        }
    }
    //======einmal beim Seiten aufruf:===================================================
    sammlungAuswahl.open("GET", serverPfad+"/sammlungAuswahl");
    sammlungAuswahl.send();
    sammlungAnzeigen();


    //Sammlungs name setzten und anzeigen
    function setSammlungName() {
        sammlungAnzeigen();
    }

    function sammlungAnzeigen() {
        document.cookie = "aktiveSammlung=" + aktuellerSammlungName;
        sammlung.open("GET", (sammlungPfad + aktuellerSammlungName));
        sammlung.send();
    }

    function setzteKlickUmZuSkalieren() {
        //bild mit neuer größe machen
        document.getElementById("neueGrose").addEventListener("click", function () {
            var tempElement = document.getElementById("aktiveBildHistorgramm");
            var bildName = tempElement.getAttribute("name");
            var breite = document.getElementById("neuBreite").value;
            var hoehe = document.getElementById("neuHoehe").value;

            resize.open("GET", serverPfad+"/resizeBild?bildname=" + bildName + "&sammlungname=" + aktuellerSammlungName + "&breite=" + breite + "&hoehe=" + hoehe);
            resize.send();
        });

    }


    document.getElementById("uploadButtonSammlung").addEventListener("click", function () {


        var fDaten = new FormData(document.getElementById("formIDSammlung"));
        uploadSammlung.open("POST", serverPfad+"/uploadBilderSammlung");
        uploadSammlung.send(fDaten);

        startLadeBalken();


    });
    function ladeAnzeigeDivErstellen() {
        var ladeanzeige = document.createElement("DIV");
        ladeanzeige.setAttribute("id", "ladeAnzeige");
        document.body.appendChild(ladeanzeige);
    }


    //klick listener auf die Sammlung Buttons
    function addClickListener() {
        var sammlungButtonDiv = document.getElementsByClassName("bilderSammlungAuswahl")[0];
        var sammlungButtons = sammlungButtonDiv.getElementsByTagName("button");

        for (i = 0; i < sammlungButtons.length; i++) {
            sammlungButtons[i].addEventListener("click", function () {
                var name = this.getAttribute("name");


                if (name.substring(0, 6) == "delete") {
                    startLadeBalken();
                    sammlungDelete.open("GET", (serverPfad+"/sammlungDelete?sammlungname=" + name.substring(6, name.length)));
                    sammlungDelete.send();
                } else {
                    aktuellerSammlungName = name;
                    sammlungAnzeigen();
                }


            })
        }
    }

    //klick listener auf die Bilder
    function addClickListenerBild() {
        var bilderButtonDiv = document.getElementsByClassName("alleBilder")[0];
        var bilderButtons = bilderButtonDiv.getElementsByTagName("img");

        for (i = 0; i < bilderButtons.length; i++) {
            bilderButtons[i].addEventListener("click", function () {
                var tempName = this.getAttribute("name")
                bildInfo.open("GET", serverPfad+"/gibBildInfo?bildname=" + tempName + "&sammlungname=" + aktuellerSammlungName);
                bildInfo.send();

            })
        }
    }


    //Funktion um neue Sammlungen zu erstellen
    document.getElementById("neuerSammlungsButton").addEventListener("click", function () {

        var tempName = prompt("Sammlungs Name eingeben", "");
        if (tempName != "" && tempName) {
            tempName = tempName.replace(/ä/g, "ae").replace(/ö/g, "oe").replace(/ü/g, "ue").replace(/Ä/g, "Ae").replace(/Ö/g, "Oe").replace(/Ü/g, "Ue").replace(/ß/g, "ss");
            aktuellerSammlungName = tempName;
            neueSammlungErstellen.open("GET", serverPfad+"/neueSammlungErstellen?name=" + tempName);
            neueSammlungErstellen.send();
        } else {
            alert("Name der Sammlung darf nicht leer sein");
        }


    })
    function startLadeBalken() {
        ladeAnzeigeDivErstellen();
        //prozentanzeige*******************
        var interval = 800;
        document.getElementById("headerff").style.visibility = "hidden";
        document.getElementById("content").style.visibility = "hidden";
        function refresh() {
            ladeAnzeige.open("GET", serverPfad+"/holProzent");
            ladeAnzeige.send();
        }

        window.clearInterval(intervalHandle);
        intervalHandle = setInterval(refresh, interval);
    }

    function endLadeBalken() {
        document.getElementById("content").style.visibility = "visible";
        document.getElementById("headerff").style.visibility = "visible";
        window.clearInterval(intervalHandle);
        document.getElementById("ladeAnzeige").remove();
    }

};
	
	