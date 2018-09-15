/**
 * Access a cookie value by name, or return null
 * @param cookieName
 */
function getCookie(cookieName) {
    let cookies = decodeURIComponent(document.cookie);

    if (cookies == "") {
        console.log("No cookies present.")
        return null;
    }

    cookies = cookies.split(";");

    for (var i = 0; i < cookies.length; i++ ){
        var cookie = cookies[i];

        var tok = cookie.split('=', 2)
        var name = tok[0].trim();
        var value = tok[1].trim();

        if (name === cookieName) {
            return value;
        }
    }

    return null;
}