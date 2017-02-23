function mosaikJS() {
	//var serverPfad = "http://localhost:4242";
	var serverPfad="";
    var aktuellerSammlungName = "SammlungTest";
    var sammlungPfad = serverPfad+"/sammlungAnzeige?sammlungname=";
    var sammlung = new XMLHttpRequest(); // anzeigen der sammlungsBidler
    var sammlungAuswahl = new XMLHttpRequest(); //anzeigen aller sammlung zum auswäheln 
    var mosaikEinstellungGenBilder = new XMLHttpRequest(); //das zu genierende mosaik kann eingstellt werden //mit parametern
    var mosaikEinstellungPoolAuswahl = new XMLHttpRequest();// zeigt alle Pools an von diesesm User
    var mosaikShowSammlung = new XMLHttpRequest();//alle Mosaik bidler anzeigen
    var machMosaik = new XMLHttpRequest(); // erstelle mosaik mit ausgewählten parametern
    var zeigeMosaikBild = new XMLHttpRequest(); // mosaik bild in groß angezeigt
    var ladeAnzeige = new XMLHttpRequest(); // lade anzeige lade balken
    var holProzent = new XMLHttpRequest(); // holt sich dei aktuellen Prozent zahl aus go
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
    zeigeMosaikBild.onreadystatechange = function () {
        if (zeigeMosaikBild.readyState === 4 && zeigeMosaikBild.status === 200) {
            document.getElementById("mosaikUIanzeige").innerHTML = zeigeMosaikBild.responseText;
            addClickListenerGroesesMosaikBild();
        }
    }
    machMosaik.onreadystatechange = function () {
        if (machMosaik.readyState === 4 && machMosaik.status === 200) {
            if (machMosaik.responseText == "zuWenigKacheln") {

                window.clearInterval(intervalHandle);
                document.getElementById("ladeAnzeige").remove();
                document.getElementById("headerff").style.visibility = "visible";
                document.getElementById("content").style.visibility = "visible";
                alert("Der Pool hat zu wenig Kacheln");
                return;
            }
            document.getElementById("headerff").style.visibility = "visible";
            document.getElementById("content").style.visibility = "visible";
            window.clearInterval(intervalHandle);
            document.getElementById("ladeAnzeige").remove();
            mosaikUIclear();
            erstelleFehlendeDivsGenTab();
            erstelleFehlendeDivsMosaikSammlung();
            mosaikShowSammlung.open("GET", serverPfad+"/mosaikShowSammlung");
            mosaikShowSammlung.send();

        }
    }
    mosaikEinstellungGenBilder.onreadystatechange = function () {
        if (mosaikEinstellungGenBilder.readyState === 4 && mosaikEinstellungGenBilder.status === 200) {
            mosaikUIclear();
            document.getElementById("mosaikUIanzeige").innerHTML = mosaikEinstellungGenBilder.responseText;
            mosaikEinstellungPoolAuswahl.open("GET", serverPfad+"/gibAllePoolsMosaik");
            mosaikEinstellungPoolAuswahl.send();


        }
    }
    mosaikEinstellungPoolAuswahl.onreadystatechange = function () {
        if (mosaikEinstellungPoolAuswahl.readyState === 4 && mosaikEinstellungPoolAuswahl.status === 200) {
            document.getElementById("mosaikGenEinstellungenPoolAuswahl").innerHTML = mosaikEinstellungPoolAuswahl.responseText;
            setzteListenerEinstellungenGenMosaik();
        }
    }
    sammlungAuswahl.onreadystatechange = function () {
        if (sammlungAuswahl.readyState === 4 && sammlungAuswahl.status === 200) {
            document.getElementById("bilderSammlungAuswahl").innerHTML = sammlungAuswahl.responseText;

            addClickListener();
        }
    }
    mosaikShowSammlung.onreadystatechange = function () {
        if (mosaikShowSammlung.readyState === 4 && mosaikShowSammlung.status === 200) {
            document.getElementById("sammlungAnzeige").innerHTML = mosaikShowSammlung.responseText;
            addClickListenerBildMosaike();
        }
    }
    //======einmal beim Seiten aufruf:===================================================
    sammlungAuswahl.open("GET", serverPfad+"/sammlungAuswahlBeiMosaikAnzeige");
    sammlungAuswahl.send();
    sammlungAnzeigen();

    document.getElementById("mosaikBildGenTab").addEventListener("click", function () {

        mosaikUIclear();
        erstelleFehlendeDivsGenTab();
        sammlungAuswahl.open("GET", serverPfad+"/sammlungAuswahlBeiMosaikAnzeige");
        sammlungAuswahl.send();
        sammlungAnzeigen();


    });
    document.getElementById("mosaikShowSammlungTab").addEventListener("click", function () {

        mosaikUIclear();
        erstelleFehlendeDivsMosaikSammlung();
        mosaikShowSammlung.open("GET", "/mosaikShowSammlung");
        mosaikShowSammlung.send();

    });


    //Pool name setzten und anzeigen
    function setSammlungName() {
        sammlungAnzeigen();
    }

    function sammlungAnzeigen() {
        document.cookie = "aktiveSammlung=" + aktuellerSammlungName;
        sammlung.open("GET", (sammlungPfad + aktuellerSammlungName));
        sammlung.send();
    }

    //klick listener auf die Sammlung Buttons
    function addClickListener() {
        var sammlungButtonDiv = document.getElementsByClassName("bilderSammlungAuswahl")[0];
        var sammlungButtons = sammlungButtonDiv.getElementsByTagName("button");

        for (i = 0; i < sammlungButtons.length; i++) {
            sammlungButtons[i].addEventListener("click", function () {
                var name = this.getAttribute("name");
                aktuellerSammlungName = name;
                sammlungAnzeigen();
            })
        }
    }

    function addClickListenerGroesesMosaikBild() {
        document.getElementById("mosaikBild").addEventListener("click", function () {
            mosaikUIclear();
            erstelleFehlendeDivsMosaikSammlung();
            mosaikShowSammlung.open("GET", serverPfad+"/mosaikShowSammlung");
            mosaikShowSammlung.send();

        })
        document.getElementById("backMosaikAnsicht").addEventListener("click", function () {
            mosaikUIclear();
            erstelleFehlendeDivsMosaikSammlung();
            mosaikShowSammlung.open("GET", serverPfad+"/mosaikShowSammlung");
            mosaikShowSammlung.send();

        })

    }

    function addClickListenerBildMosaike() {
        var bilderButtonDiv = document.getElementsByClassName("alleBilder")[0];
        var bilderButtons = bilderButtonDiv.getElementsByTagName("img");

        for (i = 0; i < bilderButtons.length; i++) {
            bilderButtons[i].addEventListener("click", function () {
                var tempName = this.getAttribute("name")
                zeigeMosaikBild.open("GET", serverPfad+"/infoMosaikBild?nameDB=" + tempName);
                zeigeMosaikBild.send();

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
                mosaikEinstellungGenBilder.open("GET", serverPfad+"/mosaikEinstellungGenBilder?bildname=" + tempName);
                mosaikEinstellungGenBilder.send();

            })
        }
    }

    function setzteListenerEinstellungenGenMosaik() {
        document.getElementById("anzahl").disabled = true;
        document.getElementById("goBack").addEventListener("click", function () {
            mosaikUIclear();
            erstelleFehlendeDivsGenTab();
            sammlungAuswahl.open("GET", serverPfad+"/sammlungAuswahlBeiMosaikAnzeige");
            sammlungAuswahl.send();
            sammlungAnzeigen();
        });
        document.getElementById("macheMosaikButton").addEventListener("click", function () {
            //Sammlen der Variablen
            var bildNameDB = document.getElementById("showAktiveMosaikVorlage").getAttribute("name")
            var bildName = document.getElementById("histoBild").getAttribute("name")
            var curr = parseInt(document.getElementById("anzahl").value);
            var kachelVerwendungMehrmals = (document.getElementById("einmal").checked == true) ? false : true;
            var nBesteKacheln = (document.getElementById("amBesten").checked == true) ? 1 : curr;
            var currPool = "";


            var allePools = document.getElementsByClassName("poolsZurAuswahl");

            for (i = 0; i < allePools.length; i++) {
                var tempName = allePools[i].getAttribute("id");
                if (document.getElementById(tempName).checked == true) {
                    currPool = allePools[i].getAttribute("id");
                }
            }
            if (currPool == "") {
                alert("Keinen Pool gewählt");
                return;
            }
            var httpString = serverPfad+"/machMosaik";

            console.log(bildNameDB);
            httpString += "?mosaikVorlageNameDB=" + bildNameDB;
            httpString += "&mosaikVorlageName=" + bildName;
            httpString += "&kachelVerwendungMehrmals=" + kachelVerwendungMehrmals;
            httpString += "&nBesteKacheln=" + nBesteKacheln;
            httpString += "&poolName=" + currPool;
            machMosaik.open("GET", httpString);
            machMosaik.send();
            ladeAnzeigeDivErstellen();

            //prozentanzeige*******************
            var interval = 800;

            function refresh() {
                ladeAnzeige.open("GET", serverPfad+"/holProzent");
                ladeAnzeige.send();
            }

            window.clearInterval(intervalHandle);
            intervalHandle = setInterval(refresh, interval);


        });
        document.getElementById("amBesten").addEventListener("click", function () {
            document.getElementById("anzahl").disabled = true;

        })


        document.getElementById("nBeste").addEventListener("click", function () {
            document.getElementById("anzahl").disabled = false;

        })
    }

    function mosaikUIclear() {
        document.getElementById("mosaikUIanzeige").remove();
        var mosaikUIanzeige = document.createElement("DIV");
        mosaikUIanzeige.setAttribute("id", "mosaikUIanzeige");
        document.getElementById("content").appendChild(mosaikUIanzeige);
    }

    function erstelleFehlendeDivsGenTab() {
        var bilderSammlung = document.createElement("DIV");
        var bilderSammlungAuswahl = document.createElement("DIV");
        var mosaikAnzeigeSammlung = document.createElement("DIV");
        var sammlungAnzeige = document.createElement("DIV");
        bilderSammlung.setAttribute("id", "bilderSammlung");
        bilderSammlungAuswahl.setAttribute("id", "bilderSammlungAuswahl");
        bilderSammlungAuswahl.setAttribute("class", "bilderSammlungAuswahl");
        mosaikAnzeigeSammlung.setAttribute("id", "mosaikAnzeigeSammlung");
        sammlungAnzeige.setAttribute("id", "sammlungAnzeige");
        bilderSammlung.appendChild(bilderSammlungAuswahl);
        mosaikAnzeigeSammlung.appendChild(sammlungAnzeige);
        document.getElementById("mosaikUIanzeige").appendChild(bilderSammlung);
        document.getElementById("mosaikUIanzeige").appendChild(mosaikAnzeigeSammlung);
    }

    function ladeAnzeigeDivErstellen() {
        var ladeanzeige = document.createElement("DIV");
        ladeanzeige.setAttribute("id", "ladeAnzeige");
        document.body.appendChild(ladeanzeige);
        document.getElementById("headerff").style.visibility = "hidden";
        document.getElementById("content").style.visibility = "hidden";
    }

    function erstelleFehlendeDivsMosaikSammlung() {
        var mosaikFertigAnzeigeSammlung = document.createElement("DIV");
        var sammlungAnzeige = document.createElement("DIV");
        mosaikFertigAnzeigeSammlung.setAttribute("id", "mosaikFertigAnzeigeSammlung");
        sammlungAnzeige.setAttribute("id", "sammlungAnzeige");
        mosaikFertigAnzeigeSammlung.appendChild(sammlungAnzeige);
        document.getElementById("mosaikUIanzeige").appendChild(mosaikFertigAnzeigeSammlung);
    }


}