function GetDomain() {
    var e = window.XMLHttpRequest ? new XMLHttpRequest : new ActiveXObject("Microsoft.XMLHTTP");
    e.responseType = "json", e.onreadystatechange = function () {
        4 == e.readyState && 200 == e.status && SetDomain(e.response)
    }, e.open("GET", "/get_domain?t=" + Math.random(), !0), e.send()
}
GetDomain();
function getcurrentdomain() {
    var myselect = document.getElementById('domains');
    index = myselect.selectedIndex;
    if (index == undefined || index < 0) {
        index = 0
    }
    if (myselect.options != undefined) {
        localStorage.setItem("domain", myselect.options[index].value)
        return myselect.options[index].value;
    }
    else {
        return ""
    }

}
function GetSubDomain() {
    if (key = localStorage.getItem("key"), null != key && 1 != confirm("获取新的子域名后将会丢失 " + key + "，请注意保存")) return !1;
    var e = window.XMLHttpRequest ? new XMLHttpRequest : new ActiveXObject("Microsoft.XMLHTTP");
    e.responseType = "json", e.onreadystatechange = function () {
        4 == e.readyState && 200 == e.status && (localStorage.setItem('key', e.response.domain), localStorage.setItem('token', e.response.token), document.getElementById("myDomain").innerHTML = e.response.domain, document.getElementById("token").innerHTML = e.response.token, GetRecords())
    }, e.open("POST", "/new_gen",true),e.setRequestHeader("Content-type","application/x-www-form-urlencoded"),e.send("domain=" + getcurrentdomain())
}
function SetDomain(obj) {
    options = "";
    select = 0;
    for (let i = 0; i < obj.length; i++) {
        //console.log(obj[i]);
        if (localStorage.getItem("domain") != null && obj[i].indexOf(localStorage.getItem("domain")) != -1) {
            options += "<option value='" + obj[i] + "' selected>" + obj[i] + "</option>";
            select = 1;
        }
        else {
            options += "<option value='" + obj[i] + "'>" + obj[i] + "</option>";
        }

    }
    document.getElementById("domains").innerHTML = options;
}
function GetRecords() {
    if (localStorage.getItem("token") == null || localStorage.getItem("domain") == null) {
        alert("Get Domain First!!")
        return;
    }
    var n = window.XMLHttpRequest ? new XMLHttpRequest : new ActiveXObject("Microsoft.XMLHTTP");
    n.onreadystatechange = function () {
        if (4 == n.readyState && 200 == n.status) {
            var e = n.responseText;
            if ("" == e || null == e || "null" == e) ktable = 'NoData', document.getElementById("myRecords").innerHTML = ktable;
            else {
                obj = JSON.parse(e), table = '';
                for (var t = Object.keys(obj).length - 1; t >= (0 < Object.keys(obj).length - 10 ? Object.keys(obj).length - 10 : 0); t--) table = table + "<tr><td>" + t + "</td><td>" + obj[t].subdomain + "</td><td>" + obj[t].ip + "</td><td>" + obj[t].time + "</td></tr>";
                document.getElementById("myRecords").innerHTML = table
            }
        }
    }, n.open("POST", "/" + localStorage.getItem("token"), true), n.setRequestHeader("Content-type", "application/x-www-form-urlencoded"), n.send("domain=" + localStorage.getItem("domain"))
}
key = localStorage.getItem("key");
token = localStorage.getItem("token");

if (key != null && token != null) {
    document.getElementById("myDomain").innerHTML = key;
    document.getElementById("token").innerHTML = token;
    GetRecords();
} else {
    //GetDomain();
    //GetSubDomain();
}
document.getElementById("myDomain").addEventListener('click', async event => {
    if (!navigator.clipboard) {
        // Clipboard API not available
        return
    }
    const text = event.target.innerText
    try {
        await navigator.clipboard.writeText(text)
        //event.target.textContent = 'Copied to clipboard'
    } catch (err) {
        console.error('Failed to copy!', err)
    }
})
document.getElementById("token").addEventListener('click', async event => {
    if (!navigator.clipboard) {
        // Clipboard API not available
        return
    }
    const text = event.target.innerText
    try {
        await navigator.clipboard.writeText(text)
        //event.target.textContent = 'Copied to clipboard'
    } catch (err) {
        console.error('Failed to copy!', err)
    }
})