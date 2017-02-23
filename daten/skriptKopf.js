window.addEventListener("load", function () {
//das sind die Händler die fast für alle Seiten gelten müssen
	//var serverPfad = "http://localhost:4242";
	var serverPfad="";
	var poolTab = new XMLHttpRequest();
	var mosaikTab = new XMLHttpRequest();
	var basisBildTab = new XMLHttpRequest();
	var logout = new XMLHttpRequest();
	var remove = new XMLHttpRequest();
	var ladeAnzeige= new XMLHttpRequest();
	var holProzent = new XMLHttpRequest();
	//lade anzeige intervall
	var intervalHandle;
	//beim erstene aufruf wird die Sammlungs auswahl angezeigt
	sammlungJS();
	//URL leiste wird ohne Reload verändert
	if ('replaceState' in history) {
		history.replaceState(null, document.title, serverPfad+"/tessera");
	}

	logout.onreadystatechange = function () {
		if (logout.readyState === 4 && logout.status === 200) {
			//weiterleitungen auf die index seite
			window.location.href = serverPfad+"/index";
		}
	}
	remove.onreadystatechange = function () {
		if (remove.readyState === 4 && remove.status === 200) {
			window.clearInterval(intervalHandle);
			//weiterleitungen auf die Index seite nachdem amn alles gelöscht hat
			window.location.href = serverPfad+"/index";
		}
	}

//hier kommt jetzt die auswahl der drei tabs wobei jede ihr eigenes js hat
	mosaikTab.onreadystatechange = function () {
		if (mosaikTab.readyState === 4 && mosaikTab.status === 200) {

			document.getElementById("content").innerHTML= mosaikTab.responseText;
			mosaikJS();

		}
	}
	basisBildTab.onreadystatechange = function () {
		if (basisBildTab.readyState === 4 && basisBildTab.status === 200) {
			document.getElementById("content").innerHTML= basisBildTab.responseText;
			sammlungJS();
		}
	}
	poolTab.onreadystatechange = function () {
		if (poolTab.readyState === 4 && poolTab.status === 200) {
			document.getElementById("content").innerHTML= poolTab.responseText;
			poolJS();
		}
	}
//klicklistener auf die Kopf zeile
	document.getElementById("mosaikTabButton").addEventListener("click", function() {
		mosaikTab.open("GET", serverPfad+"/mosaikTab");
		mosaikTab.send();

	});
	document.getElementById("basisBildTabButton").addEventListener("click", function() {
		basisBildTab.open("GET", serverPfad+"/basisBildTab");
		basisBildTab.send();

	});
	document.getElementById("poolTabButton").addEventListener("click", function() {
		poolTab.open("GET", serverPfad+"/poolTab");
		poolTab.send();

	});
	document.getElementById("deleteButton").addEventListener("click", function() {
		var flag =window.confirm("Sicher das Sie alles Löschen wollen ?");
		if(flag){
			ladeAnzeigeDivErstellen();
			//prozentanzeige*******************
			var interval = 800;

			function refresh() {
				ladeAnzeige.open("GET", serverPfad+"/holProzent");
				ladeAnzeige.send();
			}

			window.clearInterval(intervalHandle);
			intervalHandle = setInterval(refresh, interval);
			remove.open("GET", serverPfad+"/delete");
			remove.send();
		}


	});
	document.getElementById("logoutButton").addEventListener("click", function() {
		logout.open("GET", serverPfad+"/logout");
		logout.send();

	});
	//lade anzeige für das löschen des users
	function ladeAnzeigeDivErstellen(){
		var ladeanzeige = document.createElement("DIV");
		ladeanzeige.setAttribute("id","ladeAnzeige");
		document.body.appendChild(ladeanzeige);
	}

});