function poolJS() {
	//var serverPfad = "http://localhost:4242";
	var serverPfad="";
    var aktuellerPoolName = "PoolTest";
    var poolPfad = serverPfad+"/poolAnzeige?sammlungname=";
    var pool = new XMLHttpRequest(); // pool anzeigen
    var neuerPoolErstellen = new XMLHttpRequest(); // neune Pools erstellen mit parametern
    var poolsAuswahl = new XMLHttpRequest(); // auswäheln ziwshcen den pools
    var poolDelete = new XMLHttpRequest(); // lsöchen einens Pools
    var uploadPool = new XMLHttpRequest();// hocladen von bidlern
    var ladeAnzeige = new XMLHttpRequest();//lade Anzeige 
    var holProzent = new XMLHttpRequest();// holt sich prozent zahl aus go
    var intervalHandle;
    pool.onreadystatechange = function () {
        if (pool.readyState === 4 && pool.status === 200) {
            document.getElementById("poolAnzeige").innerHTML = pool.responseText;
            var poolArt = document.getElementById("poolart");
            if (poolArt.getAttribute("data-poolart")) {
                if (poolArt.getAttribute("data-poolart") == "uploadpooltypeauswahl") {
                    document.getElementById("uploadForm").style.display = "block";
                } else {
                    document.getElementById("uploadForm").style.display = "none";
                }
            }

        }
    }
    poolsAuswahl.onreadystatechange = function () {
        if (poolsAuswahl.readyState === 4 && poolsAuswahl.status === 200) {
            document.getElementById("poolsAuswahl").innerHTML = poolsAuswahl.responseText;
            addClickListener();
        }
    }
    ladeAnzeige.onreadystatechange = function () {
        if (ladeAnzeige.readyState === 4 && ladeAnzeige.status === 200) {

            document.getElementById("ladeAnzeige").innerHTML = ladeAnzeige.responseText;

        }
    }
    uploadPool.onreadystatechange = function () {
        if (uploadPool.readyState === 4 && uploadPool.status === 200) {
            endLadeBalken();
            poolAnzeigen();
            poolsAuswahl.open("GET", serverPfad+"/poolAuswahl");
            poolsAuswahl.send();


        }
    }
    poolDelete.onreadystatechange = function () {
        if (poolDelete.readyState === 4 && poolDelete.status === 200) {
            if (poolDelete.responseText == "error") {
                alert("Sie haben keine Pools mehr, es wird ein TestPool erstellt.");
            }


            poolsAuswahl.open("GET",serverPfad+"/poolAuswahl");
            poolsAuswahl.send();
            poolAnzeigen();
            endLadeBalken();


        }
    }
    neuerPoolErstellen.onreadystatechange = function () {
        if (neuerPoolErstellen.readyState === 4 && neuerPoolErstellen.status === 200) {

            if ("doppelterEintrag" == neuerPoolErstellen.responseText) {
                alert("Pool exestiert bereits");
                endLadeBalken();
                return;
            }

            poolsAuswahl.open("GET", serverPfad+"/poolAuswahl");
            poolsAuswahl.send();


            poolAnzeigen();
            endLadeBalken();
        }

    }


    //======einmal beim Seiten aufruf:===================================================
    poolsAuswahl.open("GET", serverPfad+"/poolAuswahl");
    poolsAuswahl.send();
    poolAnzeigen();


    //Pool name setzten und anzeigen
    function setPoolName() {
        poolAnzeigen();
    }

    function poolAnzeigen() {
        document.cookie = "aktivePool=" + aktuellerPoolName;
        pool.open("GET", (poolPfad + aktuellerPoolName));
        pool.send();
    }

    //klick listener auf die Pool Buttons
    function addClickListener() {
        var poolButtonDiv = document.getElementsByClassName("poolsAuswahl")[0];
        var poolButtons = poolButtonDiv.getElementsByTagName("button");

        for (i = 0; i < poolButtons.length; i++) {
            poolButtons[i].addEventListener("click", function () {
                var name = this.getAttribute("name");

                if (name.substring(0, 6) == "delete") {
                    startLadeBalken();
                    poolDelete.open("GET", (serverPfad+"/poolDelete?poolname=" + name.substring(6, name.length)));
                    poolDelete.send();
                } else {
                    aktuellerPoolName = name;
                    poolAnzeigen();
                }


            })
        }
    }

    document.getElementById("uploadButtonPool").addEventListener("click", function () {
        var fDaten = new FormData(document.getElementById("formIDPool"));

        uploadPool.open("POST", serverPfad+"/gridMultiForm");
        uploadPool.send(fDaten);
        startLadeBalken();
    });


    //Funktion um neue Pools zu erstellen
    document.getElementById("neuerPoolButton").addEventListener("click", function () {
        var curr = parseInt(document.getElementById("anzahl").value);
        var min = parseInt(document.getElementById("anzahl").min);
        var max = parseInt(document.getElementById("anzahl").max);
        if (curr < min || curr > max) {
            alert("Anzal muss minimal " + min + " und maximal " + max + " sein.");
            return;
        }

        var tempName = prompt("Pool Name eingeben", "");
        if (tempName != "" && tempName != null && tempName.length < 16) {
            tempName = tempName.replace(/ä/g, "ae").replace(/ö/g, "oe").replace(/ü/g, "ue").replace(/Ä/g, "Ae").replace(/Ö/g, "Oe").replace(/Ü/g, "Ue").replace(/ß/g, "ss");
            aktuellerPoolName = tempName;
            var httpString = serverPfad+ "/neuerPoolErstellen?name=" + tempName;
            var helligkeit = parseInt(document.getElementById("helligkeit").value);
            var farbig = (document.getElementById("einfarbig").checked == true) ? "einfarbig" : "farbverlauf";
            var bildergroesse = (document.getElementById("25x25").checked == true) ? 25 : (document.getElementById("50x50").checked == true) ? 50 : 10;

            if (document.getElementById("uploadpooltypeauswahl").checked == true) {
                httpString += "&pooltypeauswahl=uploadpooltypeauswahl&bildergroesse=" + bildergroesse;
            } else {
                httpString += "&pooltypeauswahl=generatepooltypeauswahl&anzahl=" + curr + "&helligkeit=" + helligkeit + "&farbig=" + farbig + "&bildergroesse=" + bildergroesse;
            }
            neuerPoolErstellen.open("GET", httpString);
            neuerPoolErstellen.send();

            startLadeBalken();

        } else {
            alert("Name zu lang oder leer");
            return;
        }


    })
    function ladeAnzeigeDivErstellen() {
        var ladeanzeige = document.createElement("DIV");
        ladeanzeige.setAttribute("id", "ladeAnzeige");
        document.body.appendChild(ladeanzeige);
    }

    document.getElementById("uploadpooltypeauswahl").addEventListener("click", function () {
        document.getElementById("anzahl").disabled = true;
        document.getElementById("helligkeit").disabled = true;
        document.getElementById("einfarbig").disabled = true;
        document.getElementById("farbverlauf").disabled = true;

    })


    document.getElementById("generatepooltypeauswahl").addEventListener("click", function () {
        document.getElementById("anzahl").disabled = false;
        document.getElementById("helligkeit").disabled = false;
        document.getElementById("einfarbig").disabled = false;
        document.getElementById("farbverlauf").disabled = false;

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
        document.getElementById("headerff").style.visibility = "visible";
        document.getElementById("content").style.visibility = "visible";
        window.clearInterval(intervalHandle);
        if (document.getElementById("ladeAnzeige")) {
            document.getElementById("ladeAnzeige").remove();
        }

    }
};
