package html

import "testing"

func TestInnerScript(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	doc, err := Parse([]byte("<html><body><div><h1></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	h1 := doc.Root().FirstChild().FirstChild().FirstChild()
	h1.SetInnerHtml("<script>if (suppressReviews !== 'true' && app == 'PRR') { ok = true; }</script>")
	if h1.String() != "<h1><script>if (suppressReviews !== 'true' && app == 'PRR') { ok = true; }</script></h1>" {
		t.Error("script does not match")
	}
	doc.Free()
}

func TestInnerScript2(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)
	script := `<script>try {
var productNAPage = "",
suppressReviews = "false";
var bvtoken = MACYS.util.Cookie.get("BazaarVoiceToken","GCs");
//bvtoken=bvtoken.substring(0,bvtoken.length-1);
$BV.configure("global", {
userToken: bvtoken,
productId: '531726',
submissionUI: 'LIGHTBOX',
submissionContainerUrl: window.location.href,
allowSamePageSubmission: true,
doLogin: function(callback, success_url) {
MACYS.util.Cookie.set("FORWARDPAGE_KEY",success_url);
window.location = 'https://www.macys.com/signin/index.ognc?fromPage=pdpReviews';
},
doShowContent: function(app, dc, sub, sr) {
if (suppressReviews !== 'true' && app == "PRR") {
MACYS.pdp.showReviewsTab();
} else if (productNAPage !== 'true' && app == "QA") {
MACYS.pdp.showQATab();
}
}
});
if (suppressReviews !== 'true') {
$BV.ui('rr', 'show_reviews', {
});
}
$BV.ui("qa", "show_questions", {
subjectType: 'product'
});
} catch ( e ) { }</script>`

	doc, err := Parse([]byte("<html><body><div><h1></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	h1 := doc.Root().FirstChild().FirstChild().FirstChild()
	h1.SetInnerHtml(script)
	if h1.String() != "<h1>"+script+"</h1>" {
		t.Error("script does not match")
	}
	doc.Free()
}

func TestInsertMyselfBefore(t *testing.T) {
	input := `<html>
<head>
<title> Title </title>
</head>
<body>
<div id="header"></div>
<h1> Welcome to Tritium Tester </h1>
</body>
</html>
`
	doc, err := Parse([]byte(input), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	divs, _ := doc.Search("//div")
	if len(divs) != 1 {
		t.Error("should have 1 div")
		return
	}

	div := divs[0]
	div.InsertBefore(div)

	expected := `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<title> Title </title>
</head>
<body>
<div id="header"></div>
<h1> Welcome to Tritium Tester </h1>
</body>
</html>
`
	if expected != doc.String() {
		t.Error("doc is not expected:\n", doc.String(), "\n", expected)
	}
	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestInsertMyselfAfter(t *testing.T) {
	input := `<html>
<head>
<title> Title </title>
</head>
<body>
<div id="header"></div>
<h1> Welcome to Tritium Tester </h1>
</body>
</html>
`
	doc, err := Parse([]byte(input), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	divs, _ := doc.Search("//div")
	if len(divs) != 1 {
		t.Error("should have 1 div")
		return
	}

	div := divs[0]
	div.InsertAfter(div)

	expected := `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<title> Title </title>
</head>
<body>
<div id="header"></div>
<h1> Welcome to Tritium Tester </h1>
</body>
</html>
`
	if expected != doc.String() {
		t.Error("doc is not expected:\n", doc.String(), "\n", expected)
	}
	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestAddMyselfChild(t *testing.T) {
	input := `<html>
<head>
<title> Title </title>
</head>
<body>
<div id="header"></div>
<h1> Welcome to Tritium Tester </h1>
</body>
</html>
`
	doc, err := Parse([]byte(input), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	divs, _ := doc.Search("//div")
	if len(divs) != 1 {
		t.Error("should have 1 div")
		return
	}

	div := divs[0]
	div.AddChild(div)

	expected := `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<title> Title </title>
</head>
<body>
<div id="header"></div>
<h1> Welcome to Tritium Tester </h1>
</body>
</html>
`
	if expected != doc.String() {
		t.Error("doc is not expected:\n", doc.String(), "\n", expected)
	}
	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestRemoveMeRemoveParent(t *testing.T) {
	input := `<html>
<head>
<title> Title </title>
</head>
<body>
<div id="header"><h1> Welcome to Tritium Tester </h1></div>
</body>
</html>
`
	doc, err := Parse([]byte(input), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	divs, _ := doc.Search("//div")
	if len(divs) != 1 {
		t.Error("should have 1 div")
		return
	}

	div := divs[0]
	h1 := div.FirstChild()
	nodes, _ := h1.Search("..")
	h1.Remove()
	nodes, _ = h1.Search("..")
	if len(nodes) != 1 {
		t.Error("removed node should have a parent , i.e. its document")
	}
	nodes[0].Remove()
	doc.Free()
	CheckXmlMemoryLeaks(t)
}
