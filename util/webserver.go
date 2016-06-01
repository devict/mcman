package util

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var output_channel chan string

type WebUser struct {
	Username string
	Password string
}

var session_store = sessions.NewCookieStore([]byte("super_secret_secret :D"))

func StartServer(ch chan string) {
	output_channel = ch
	_, err := os.Stat("mapcrafter/index.html")
	if err == nil {
		// Looks like mapcrafter is present
		output_channel <- "* Mapcrafter Directory is Present, routing to /\n"
		fs := http.FileServer(http.Dir("mapcrafter"))
		http.Handle("/", fs)
	}
	http.HandleFunc("/assets/", getAsset)
	http.HandleFunc("/admin/", serveMcMan)
	output_channel <- "* Admin site running at /admin/\n"
	http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}

func serveMcMan(w http.ResponseWriter, r *http.Request) {
	output := ""
	output_channel <- fmt.Sprint("HTTP Request: (", r.Method, ") ", r.URL, "\n")

	the_path := r.URL.Path
	the_path = strings.TrimPrefix(strings.ToLower(the_path), "/admin")

	if strings.HasPrefix(the_path, "/login") || the_path == "/" {
		output = getHTML(loginScreen, w, r)
	} else if strings.HasPrefix(the_path, "/dologin") {
		output = getHTML(doLogin, w, r)
	} else if strings.HasPrefix(the_path, "/api/") {
		output = getAuthJSON(serveAPI, w, r)
	}
	fmt.Fprintf(w, output)
}

func serveAPI(w http.ResponseWriter, r *http.Request) string {
	_, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	//body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// What are we doing with this request?
	the_path := strings.TrimPrefix(strings.ToLower(r.URL.Path), "/admin")
	output_string := ""
	if strings.HasPrefix(the_path, "/api") {
		the_path = strings.TrimPrefix(the_path, "/api")
		if strings.HasPrefix(the_path, "/v1") {
			the_path = strings.TrimPrefix(the_path, "/v1")
			if strings.HasPrefix(the_path, "/whitelist") {
				output_string = handleWhitelist(r)
			} else if strings.HasPrefix(the_path, "/ops") {
				output_string = handleOps(r)
			} else if strings.HasPrefix(the_path, "/stop") {
				DoStopServer()
			} else if strings.HasPrefix(the_path, "/init") {
				v := WebUser{"br0xen", "asdf"}
				c.model.updateWebUser(&v)
			}
		}
	}
	return output_string
}

/* JSON Functions */
func getJSON(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	w.Header().Set("Content-Type", "application/json")
	return mid(w, r)
}

func getAuthJSON(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	session, _ := session_store.Get(r, "mc_man_session")

	// Vaildate that the session is authenticated
	if val, ok := session.Values["is_logged_in"].(string); ok {
		switch val {
		case "":
			w.WriteHeader(403)
			w.Header().Set("Content-Type", "application/json")
			return "{\"status\":\"error\"}"
		default:
			w.Header().Set("Content-Type", "application/json")
			return mid(w, r)
		}
	} else {
		w.WriteHeader(403)
		w.Header().Set("Content-Type", "application/json")
		return "{\"status\":\"error\"}"
	}
	return ""
}

func handleOps(r *http.Request) string {
	if r.Method == "GET" {
		return getOps()
	} else if r.Method == "POST" {
		// Add posted user to Ops
	}
	return ""
}

func handleWhitelist(r *http.Request) string {
	if r.Method == "GET" {
		return getWhitelist()
	} else if r.Method == "POST" {
		// Add posted user to whitelist
	}
	return ""
}

/* HTML Functions */
func getAuthHTML(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	session, _ := session_store.Get(r, "mc_man_session")

	// Vaildate that the session is authenticated
	if val, ok := session.Values["is_logged_in"].(string); ok {
		switch val {
		case "":
			http.Redirect(w, r, "/admin/login", http.StatusFound)
		default:
			return mid(w, r)
		}
	} else {
		http.Redirect(w, r, "/admin/login", http.StatusFound)
	}
	return ""
}

func getHTML(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	return fmt.Sprintf("%s%s%s", htmlHeader("Minecraft Manager"), mid(w, r), htmlFooter())
}

func loginScreen(w http.ResponseWriter, r *http.Request) string {
	return `
	<form class="pure-form" action="/admin/doLogin" method="POST">
		<fieldset>
			<input type="text" name="username" placeholder="username">
			<input type="password" name="password" placeholder="password">
			<button type="submit" class="pure-button pure-button-primary button-success">sign in</button>
		</fieldset>
	</form>
`
}

func doLogin(w http.ResponseWriter, r *http.Request) string {
	ret := "Do Login<br />"
	login_user := r.FormValue("username")
	login_pass := r.FormValue("password")
	lu := c.model.getWebUser(login_user)
	// Set the Cookie

	if login_pass == lu.Password {
		session, _ := session_store.Get(r, "mc_man_session")
		session.Values["is_logged_in"] = login_user
		session.Save(r, w)

		ret = ret + "Logged In!"
	}

	return ret
}

func htmlHeader(title string) string {
	head := `
<!DOCTYPE html>
<html>
	<head>
		<title>`
	head += title
	head += `
		</title>
		<link rel="stylesheet" href="/assets/css/pure.css">
		<link rel="stylesheet" href="/assets/css/mc_man.css">
<!--[if lte IE 8]>
    <link rel="stylesheet" href="http://yui.yahooapis.com/pure/0.6.0/grids-responsive-old-ie-min.css">
<![endif]-->
<!--[if gt IE 8]><!-->
    <link rel="stylesheet" href="http://yui.yahooapis.com/pure/0.6.0/grids-responsive-min.css">
<!--<![endif]-->
	</head>
	<body>
	<div class="mcman_wrapper pure-g" id="menu">
		<div class="pure-u-1 pure-u-md-1-3">
			<div class="pure-menu">
				<a href="#" class="pure-menu-heading custom-brand">Buller Mineworks</a>
				<a href="#" class="mcman-toggle" id="toggle"><s class="bar"></s><s class="bar"></s></a>
			</div>
		</div>
		<div class="pure-u-1 pure-u-md-1-3">
			<div class="pure-menu pure-menu-horizontal mcman-can-transform">
				<ul class="pure-menu-list">
					<li class="pure-menu-item"><a href="#" class="pure-menu-link">Home</a></li>
					<li class="pure-menu-item"><a id="stop_link" href="#" class="pure-menu-link">Stop</a></li>
					<li class="pure-menu-item"><a href="#" class="pure-menu-link">About</a></li>
				</ul>
			</div>
		</div>
		<div class="pure-u-1 pure-u-md-1-3">
			<div class="pure-menu pure-menu-horizontal mcman-menu-3 mcman-can-transform">
				<ul class="pure-menu-list">
					<li class="pure-menu-item"><a href="/" class="pure-menu-link">Map</a></li>
				</ul>
			</div>
		</div>
	</div>
`
	return head
}

func htmlFooter() string {
	return `
	<script src="/assets/js/B.js"></script>
	<script src="/assets/js/mc_man.js"></script>
	</body>
</html>
`
}

/* Data Functions */
func getOps() string {
	ret := "["
	num_users := 0
	for _, op_user := range GetConfig().Ops {
		if num_users > 0 {
			ret += ","
		}
		ret += fmt.Sprint("\"", op_user, "\"")
	}
	ret += "]"
	return ret
}

func getWhitelist() string {
	ret := "["
	num_users := 0
	for _, wl_user := range GetConfig().Whitelist {
		if num_users > 0 {
			ret += ","
		}
		ret += fmt.Sprint("\"", wl_user, "\"")
	}
	ret += "]"
	return ret
}

/* Assets (Javascript/CSS) */
func getAsset(w http.ResponseWriter, r *http.Request) {
	output := ""
	output_channel <- fmt.Sprint("HTTP Request: (", r.Method, ") ", r.URL, "\n")

	the_path := strings.ToLower(r.URL.Path)
	the_path = strings.TrimPrefix(the_path, "/assets")
	if strings.HasPrefix(the_path, "/js") {
		w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
		the_path = strings.TrimPrefix(the_path, "/js")
		if strings.HasPrefix(the_path, "/b.js") {
			output = getjsB()
		} else if strings.HasPrefix(the_path, "/mc_man.js") {
			output = getjsMcMan()
		}
	} else if strings.HasPrefix(the_path, "/css") {
		w.Header().Set("Content-Type", "text/css; charset=UTF-8")
		the_path = strings.TrimPrefix(the_path, "/css")
		if strings.HasPrefix(the_path, "/pure.css") {
			output = getcssPure()
		} else if strings.HasPrefix(the_path, "/mc_man.css") {
			output = getcssMcMan()
		}
	}
	fmt.Fprintf(w, output)
}

func getjsMcMan() string {
	return `
(function(window,document){
	B('#stop_link').on('click',function(){
		var oReq = new XMLHttpRequest();
		oReq.onload = function(){console.log(this.responseText);};
		oReq.open("get", "/admin/api/v1/stop", true);
		oReq.send();
	});
	var menu=B('#menu'),
			WINDOW_CHANGE_EVENT=('onorientationchange' in window)?'orientationchange':'resize';
	function toggleHorizontal(){
		[].forEach.call(
				B('#menu .custom-can-transform'),
				function(el){
					el.classList.toggle('pure-menu-horizontal');
				}
		);
	};
	function toggleMenu(){
		if(menu[0].classList.contains('open')){
			setTimeout(toggleHorizontal,500);
		} else {
			toggleHorizontal();
		}
		menu.classList.toggle('open');
		B('#toggle').classList.toggle('x');
	};
	function closeMenu(){
		if(menu[0].classList.contains('open')){
			toggleMenu();
		}
	};
	B('#toggle').on('click', function(e){
		toggleMenu();
	});
	window.addEventListener(WINDOW_CHANGE_EVENT,closeMenu);
})(this,this.document);
`
}

func getcssMcMan() string {
	return `
button{padding:8px;}
.button-success,.button-error,.button-warning,.button-secondary{
color: white;border-radius: 4px;text-shadow: 0 1px 1px rgba(0, 0, 0, 0.2);}
.button-success{background: rgb(28, 184, 65);}
.button-error{background: rgb(202, 60, 60);}
.button-warning{background: rgb(223, 117, 20);}
.button-secondary{background: rgb(66, 184, 221);}
// Menu Stuff
.mcman-wrapper{background-color:#ffd390;margin-bottom:1em;-webkit-font-smoothing:antialiased;
height:2.1em;overflow:hidden;-webkit-transition:height 0.5s;-moz-transition:height 0.5s;
-ms-transition:height 0.5s;transition:height 0.5s;}
.mcman-wrapper.open{height:14em;}
.mcman-menu-3{text-align:right;}
.mcman-toggle{width:34px;height:34px;display:block;position:absolute;top:0;right:0;display:none;}
.mcman-toggle .bar{background-color:#777;display:block;width:20px;height:2px;border-radius:100px;
position:absolute;top:18px;right:7px;-webkit-transition:all 0.5s;-moz-transition:all 0.5s;
-ms-transition:all 0.5s;transition:all 0.5s;}
.mcman-toggle .bar:first-child{-webkit-transform:translateY(-6px);-moz-transform:translateY(-6px);
-ms-transform:translateY(-6px);transform:translateY(-6px);}
.mcman-toggle.x .bar{-webkit-transform:rotate(45deg);-moz-transform:rotate(45deg);
-ms-transform:rotate(45deg);transform:rotate(45deg);}
.mcman-toggle.x .bar:first-child{-webkit-transform:rotate(-45deg);-moz-transform:rotate(-45deg);
-ms-transform:rotate(-45deg);transform:rotate(-45deg);}
@media (max-width:47.999em){.custom-menu-3{text-align:left;}.custom-toggle{display:block;}
`
}

func getjsB() string {
	return `
function B(els, attrs) {
  // Turn 'this' into an array of passed in elements.
  function B(els) {
    if(typeof els === "string") {
      els = this.brb.create(els);
    }
    for(var i = 0; i < els.length; i++) {
      this[i] = els[i];
    }
    this.length = els.length;
//    return this;
  }

  // Map a function to all elements in 'this'
  B.prototype.map = function(callback) {
    var results = [], i = 0;
    for(;i<this.length;i++){
      results.push(callback.call(this,this[i],i));
    }
    return results;
  };

  // Foreach through all elements in 'this'
  B.prototype.forEach = function(callback) {
    this.map(callback);
    return this;
  };

  // Map a function to the first element in 'this'
  B.prototype.mapOne = function(callback) {
    var m = this.map(callback);
    return m.length > 1 ? m : m[0];
  };

  // Update css for each element in 'this'
  B.prototype.css = function(css_opt_var, css_opt_val) {
    if(typeof css_opt_var !== "string") {
      for(css_var in css_opt_var) {
        this.forEach(function(el){el.style[css_var]=css_opt_var[css_var];});
      }
      return this;
    } else {
      if(typeof css_opt_val !== "undefined") {
        return this.forEach(function(el){el.style[css_opt_var]=css_opt_val;});
      } else {
        return this.mapOne(function(el){return el.style[css_opt_var];});
      }
    }
  };

  // Update the innerText for each element in 'this'
  B.prototype.text = function(text) {
    if(typeof text !== "undefined") {
      return this.forEach(function(el){el.innerText=text;});
    } else {
      return this.mapOne(function(el){return el.innerText;});
    }
  };

  // Add a class to each element in 'this'
  B.prototype.addClass = function(classes) {
    var className = "";
    if(typeof classes !== "string") {
      for(var i=0;i<classes.length;i++) {
        className+=" "+classes[i];
      }
    } else {
      className=" "+classes;
    }
    return this.forEach(function(el){el.className+=className;});
  };

  // Remove a class from each element in 'this'
  B.prototype.removeClass = function(remove_class) {
    return this.forEach(function(el){
      var cs = el.className.split(" "), i;
      while((i=cs.indexOf(remove_class))>-1){
        cs = cs.slice(0,i).concat(cs.slice(++i));
      }
      el.className=cs.join(" ");
    });
  };

  // Set an attribute for each element in 'this'
  B.prototype.attr = function(attr,val){
    if(typeof val!=="undefined"){
      if(this[0].tagName=="INPUT" && attr.toUpperCase()=="VALUE") {
        // If we're setting the 'VALUE' then it's actually .value
        return this.forEach(function(el){
          el.value=val;
        });
      } else {
        // Otherwise use .setAttribute
        return this.forEach(function(el){
          el.setAttribute(attr,val);
        });
      }
    } else {
      // And clearing the value
      if(this[0].tagName=="INPUT" && attr.toUpperCase()=="VALUE") {
        return this.mapOne(function(el){
          return el.value;
        });
      } else {
        return this.mapOne(function(el){
          return el.getAttribute(attr);
        });
      }
    }
  };

  // Actually set a value on each element (can be done with attr too.)
  B.prototype.val = function(new_val) {
    if(typeof new_val!=="undefined"){
      return this.forEach(function(el){
        el.value = new_val;
      });
    } else {
      // Just retrieve the value for the first element
      return this.mapOne(function(el) {
        return el.value;
      });
    }
  }

  // Append an element to the DOM after each element in 'this'
  B.prototype.append = function(els) {
    this.forEach(function(parEl, i) {
      els.forEach(function(childEl) {
        if(i>0) {
          childEl=childEl.cloneNode(true);
        }
        parEl.appendChild(childEl);
      });
    });
  };

  // Prepend an element to the DOM before each element in 'this'
  B.prototype.prepend = function(els) {
    return this.forEach(function(parEl, i) {
      for(var j = els.length-1; j>-1; j--) {
        childEl=(i>0)?els[j].cloneNode(true):els[j];
        parEl.insertBefore(childEl, parEl.firstChild);
      }
    });
  };

  // Remove all elements in 'this' from the DOM
  B.prototype.remove = function() {
    return this.forEach(function(el){
      return el.parentNode.removeChild(el);
    });
  };

  // Find children that match selector
  B.prototype.children = function(selector) {
    var results = [];
    this.forEach(function(el) {
      var sub_r = el.querySelectorAll(selector);
      for(var i = 0; i < sub_r.length; i++) {
        results.push(sub_r[i]);
      }
    });
    return results;
  }

  B.prototype.firstChild = function(selector) {
    return this.children(selector)[0];
  }

  // Add an event listener to each element in 'this'
  B.prototype.on = (function(){
    // Browser compatibility...
    if(document.addEventListener) {
      return function(evt,fn) {
        return this.forEach(function(el){
          el.addEventListener(evt, fn, false);
        });
      };
    } else if(document.attachEvent) {
      return function(evt,fn) {
        return this.forEach(function(el){
          el.attachEvent("on"+evt,fn);
        });
      };
    } else {
      return function(evt, fn) {
        return this.forEach(function(el){
          el["on"+evt]=fn;
        });
      };
    }
  }());

  // Disable event listeners on elements in 'this'
  B.prototype.off = (function(){
    // Browser compatibility...
    if(document.removeEventListener) {
      return function(evt, fn) {
        return this.forEach(function(el) {
          el.removeEventListener(evt, fn, false);
        });
      };
    } else if(document.detachEvent) {
      return function(evt, fn) {
        return this.forEach(function(el) {
          el.detachEvent("on"+evt, fn);
        });
      };
    } else {
      return function(evt, fn) {
        return this.forEach(function(el){
          el["on"+evt]=null;
        });
      };
    }
  }());

  // The actual framework object, yay!
  var brb = {
    // Get an element
    get: function(selector) {
      var els;
      if(typeof selector === "string") {
        els = document.querySelectorAll(selector);
      } else if(selector.length) {
        els = selector;
      } else {
        els = [selector];
      }
      return new B(els);
    },
    // Create a new element
    create: function(tagName, attrs) {
      var el = new B([document.createElement(tagName)]);
      // Set attributes on new element
      if(attrs) {
        if(attrs.className) {
          // Classes
          el.addClass(attrs.className);
          delete attrs.classname;
        }
        if(attrs.text) {
          // Text
          el.text(attrs.text);
          delete attrs.text;
        }
        for(var key in attrs) {
          // All other Attributes
          if(attrs.hasOwnProperty(key)) {
            el.attr(key, attrs[key]);
          }
        }
      }
      return el;
    }
  };
  if(els.match) {
    var match_tags = els.match(/<([^>\s\/]*)\s?\/?>/);
  }
  if(match_tags && match_tags.length > 0) {
    // It's a 'create tag' command
    return brb.create(match_tags[1], attrs);
  } else {
    // Just search for matches
    return brb.get(els);
  }
};
`
}

func getcssPure() string {
	return `
/*!
Pure v0.6.0
Copyright 2014 Yahoo! Inc. All rights reserved.
Licensed under the BSD License.
https://github.com/yahoo/pure/blob/master/LICENSE.md
*/
/*!
normalize.css v^3.0 | MIT License | git.io/normalize
Copyright (c) Nicolas Gallagher and Jonathan Neal
*/
/*! normalize.css v3.0.2 | MIT License | git.io/normalize */html{font-family:sans-serif;-ms-text-size-adjust:
100%;-webkit-text-size-adjust:100%}body{margin:0}article,aside,details,figcaption,figure,footer,header,hgroup,
main,menu,nav,section,summary{display:block}audio,canvas,progress,video{display:inline-block;vertical-align:
baseline}audio:not([controls]){display:none;height:0}[hidden],template{display:none}a{background-color:
transparent}a:active,a:hover{outline:0}abbr[title]{border-bottom:1px dotted}b,strong{font-weight:700}dfn{
font-style:italic}h1{font-size:2em;margin:.67em 0}mark{background:#ff0;color:#000}small{font-size:80%}sub,sup{
font-size:75%;line-height:0;position:relative;vertical-align:baseline}sup{top:-.5em}sub{bottom:-.25em}img{
border:0}svg:not(:root){overflow:hidden}figure{margin:1em 40px}hr{-moz-box-sizing:content-box;box-sizing:
content-box;height:0}pre{overflow:auto}code,kbd,pre,samp{font-family:monospace,monospace;font-size:1em}
button,input,optgroup,select,textarea{color:inherit;font:inherit;margin:0}button{overflow:visible}button,
select{text-transform:none}button,html input[type=button],input[type=reset],input[type=submit]{
-webkit-appearance:button;cursor:pointer}button[disabled],html input[disabled]{cursor:default}
button::-moz-focus-inner,input::-moz-focus-inner{border:0;padding:0}input{line-height:normal}
input[type=checkbox],input[type=radio]{box-sizing:border-box;padding:0}
input[type=number]::-webkit-inner-spin-button,input[type=number]::-webkit-outer-spin-button{height:auto}
input[type=search]{-webkit-appearance:textfield;-moz-box-sizing:content-box;-webkit-box-sizing:content-box;
box-sizing:content-box}input[type=search]::-webkit-search-cancel-button,
input[type=search]::-webkit-search-decoration{-webkit-appearance:none}fieldset{border:1px solid silver;
margin:0 2px;padding:.35em .625em .75em}legend{border:0;padding:0}textarea{overflow:auto}optgroup{
font-weight:700}table{border-collapse:collapse;border-spacing:0}td,th{padding:0}.hidden,[hidden]{
display:none!important}.pure-img{max-width:100%;height:auto;display:block}.pure-g{letter-spacing:-.31em;
*letter-spacing:normal;*word-spacing:-.43em;text-rendering:optimizespeed;
font-family:FreeSans,Arimo,"Droid Sans",Helvetica,Arial,sans-serif;display:-webkit-flex;
-webkit-flex-flow:row wrap;display:-ms-flexbox;-ms-flex-flow:row wrap;-ms-align-content:flex-start;
-webkit-align-content:flex-start;align-content:flex-start}.opera-only :-o-prefocus,.pure-g{word-spacing:-.43em}
.pure-u{display:inline-block;*display:inline;zoom:1;letter-spacing:normal;word-spacing:normal;
vertical-align:top;text-rendering:auto}.pure-g [class *="pure-u"]{font-family:sans-serif}.pure-u-1,
.pure-u-1-1,.pure-u-1-2,.pure-u-1-3,.pure-u-2-3,.pure-u-1-4,.pure-u-3-4,.pure-u-1-5,.pure-u-2-5,.pure-u-3-5,
.pure-u-4-5,.pure-u-5-5,.pure-u-1-6,.pure-u-5-6,.pure-u-1-8,.pure-u-3-8,.pure-u-5-8,.pure-u-7-8,.pure-u-1-12,
.pure-u-5-12,.pure-u-7-12,.pure-u-11-12,.pure-u-1-24,.pure-u-2-24,.pure-u-3-24,.pure-u-4-24,.pure-u-5-24,
.pure-u-6-24,.pure-u-7-24,.pure-u-8-24,.pure-u-9-24,.pure-u-10-24,.pure-u-11-24,.pure-u-12-24,.pure-u-13-24,
.pure-u-14-24,.pure-u-15-24,.pure-u-16-24,.pure-u-17-24,.pure-u-18-24,.pure-u-19-24,.pure-u-20-24,
.pure-u-21-24,.pure-u-22-24,.pure-u-23-24,.pure-u-24-24{display:inline-block;*display:inline;zoom:1;
letter-spacing:normal;word-spacing:normal;vertical-align:top;text-rendering:auto}.pure-u-1-24{
width:4.1667%;*width:4.1357%}.pure-u-1-12,.pure-u-2-24{width:8.3333%;*width:8.3023%}.pure-u-1-8,
.pure-u-3-24{width:12.5%;*width:12.469%}.pure-u-1-6,.pure-u-4-24{width:16.6667%;*width:16.6357%}
.pure-u-1-5{width:20%;*width:19.969%}.pure-u-5-24{width:20.8333%;*width:20.8023%}.pure-u-1-4,.pure-u-6-24{
width:25%;*width:24.969%}.pure-u-7-24{width:29.1667%;*width:29.1357%}.pure-u-1-3,.pure-u-8-24{width:33.3333%;
*width:33.3023%}.pure-u-3-8,.pure-u-9-24{width:37.5%;*width:37.469%}.pure-u-2-5{width:40%;*width:39.969%}
.pure-u-5-12,.pure-u-10-24{width:41.6667%;*width:41.6357%}.pure-u-11-24{width:45.8333%;*width:45.8023%}
.pure-u-1-2,.pure-u-12-24{width:50%;*width:49.969%}.pure-u-13-24{width:54.1667%;*width:54.1357%}.pure-u-7-12,
.pure-u-14-24{width:58.3333%;*width:58.3023%}.pure-u-3-5{width:60%;*width:59.969%}.pure-u-5-8,.pure-u-15-24{
width:62.5%;*width:62.469%}.pure-u-2-3,.pure-u-16-24{width:66.6667%;*width:66.6357%}.pure-u-17-24{
width:70.8333%;*width:70.8023%}.pure-u-3-4,.pure-u-18-24{width:75%;*width:74.969%}.pure-u-19-24{
width:79.1667%;*width:79.1357%}.pure-u-4-5{width:80%;*width:79.969%}.pure-u-5-6,.pure-u-20-24{
width:83.3333%;*width:83.3023%}.pure-u-7-8,.pure-u-21-24{width:87.5%;*width:87.469%}.pure-u-11-12,
.pure-u-22-24{width:91.6667%;*width:91.6357%}.pure-u-23-24{width:95.8333%;*width:95.8023%}.pure-u-1,
.pure-u-1-1,.pure-u-5-5,.pure-u-24-24{width:100%}.pure-button{display:inline-block;zoom:1;line-height:normal;
white-space:nowrap;vertical-align:middle;text-align:center;cursor:pointer;-webkit-user-drag:none;
-webkit-user-select:none;-moz-user-select:none;-ms-user-select:none;user-select:none;
-webkit-box-sizing:border-box;-moz-box-sizing:border-box;box-sizing:border-box}.pure-button::-moz-focus-inner{
padding:0;border:0}.pure-button{font-family:inherit;font-size:100%;padding:.5em 1em;color:#444;
color:rgba(0,0,0,.8);border:1px solid #999;border:0 rgba(0,0,0,0);background-color:#E6E6E6;
text-decoration:none;border-radius:2px}.pure-button-hover,.pure-button:hover,.pure-button:focus{
filter:progid:DXImageTransform.Microsoft.gradient(startColorstr='#00000000', endColorstr='#1a000000',
GradientType=0);background-image:-webkit-gradient(linear,0 0,0 100%,from(transparent),
color-stop(40%,rgba(0,0,0,.05)),to(rgba(0,0,0,.1)));
background-image:-webkit-linear-gradient(transparent,rgba(0,0,0,.05) 40%,rgba(0,0,0,.1));
background-image:-moz-linear-gradient(top,rgba(0,0,0,.05) 0,rgba(0,0,0,.1));
background-image:-o-linear-gradient(transparent,rgba(0,0,0,.05) 40%,rgba(0,0,0,.1));
background-image:linear-gradient(transparent,rgba(0,0,0,.05) 40%,rgba(0,0,0,.1))}.pure-button:focus{
outline:0}.pure-button-active,.pure-button:active{
box-shadow:0 0 0 1px rgba(0,0,0,.15) inset,0 0 6px rgba(0,0,0,.2) inset;border-color:#000\9}
.pure-button[disabled],.pure-button-disabled,.pure-button-disabled:hover,.pure-button-disabled:focus,
.pure-button-disabled:active{border:0;background-image:none;
filter:progid:DXImageTransform.Microsoft.gradient(enabled=false);filter:alpha(opacity=40);-khtml-opacity:.4;
-moz-opacity:.4;opacity:.4;cursor:not-allowed;box-shadow:none}.pure-button-hidden{display:none}
.pure-button::-moz-focus-inner{padding:0;border:0}.pure-button-primary,.pure-button-selected,
a.pure-button-primary,a.pure-button-selected{background-color:#0078e7;color:#fff}.pure-form input[type=text],
.pure-form input[type=password],.pure-form input[type=email],.pure-form input[type=url],
.pure-form input[type=date],.pure-form input[type=month],.pure-form input[type=time],
.pure-form input[type=datetime],.pure-form input[type=datetime-local],.pure-form input[type=week],
.pure-form input[type=number],.pure-form input[type=search],.pure-form input[type=tel],
.pure-form input[type=color],.pure-form select,.pure-form textarea{padding:.5em .6em;display:inline-block;
border:1px solid #ccc;box-shadow:inset 0 1px 3px #ddd;border-radius:4px;vertical-align:middle;
-webkit-box-sizing:border-box;-moz-box-sizing:border-box;box-sizing:border-box}.pure-form input:not([type]){
padding:.5em .6em;display:inline-block;border:1px solid #ccc;box-shadow:inset 0 1px 3px #ddd;
border-radius:4px;-webkit-box-sizing:border-box;-moz-box-sizing:border-box;box-sizing:border-box}
.pure-form input[type=color]{padding:.2em .5em}.pure-form input[type=text]:focus,
.pure-form input[type=password]:focus,.pure-form input[type=email]:focus,.pure-form input[type=url]:focus,
.pure-form input[type=date]:focus,.pure-form input[type=month]:focus,.pure-form input[type=time]:focus,
.pure-form input[type=datetime]:focus,.pure-form input[type=datetime-local]:focus,
.pure-form input[type=week]:focus,.pure-form input[type=number]:focus,.pure-form input[type=search]:focus,
.pure-form input[type=tel]:focus,.pure-form input[type=color]:focus,.pure-form select:focus,
.pure-form textarea:focus{outline:0;border-color:#129FEA}.pure-form input:not([type]):focus{
outline:0;border-color:#129FEA}.pure-form input[type=file]:focus,.pure-form input[type=radio]:focus,
.pure-form input[type=checkbox]:focus{outline:thin solid #129FEA;outline:1px auto #129FEA}
.pure-form .pure-checkbox,.pure-form .pure-radio{margin:.5em 0;display:block}
.pure-form input[type=text][disabled],.pure-form input[type=password][disabled],
.pure-form input[type=email][disabled],.pure-form input[type=url][disabled],
.pure-form input[type=date][disabled],.pure-form input[type=month][disabled],
.pure-form input[type=time][disabled],.pure-form input[type=datetime][disabled],
.pure-form input[type=datetime-local][disabled],.pure-form input[type=week][disabled],
.pure-form input[type=number][disabled],.pure-form input[type=search][disabled],
.pure-form input[type=tel][disabled],.pure-form input[type=color][disabled],.pure-form select[disabled],
.pure-form textarea[disabled]{cursor:not-allowed;background-color:#eaeded;color:#cad2d3}
.pure-form input:not([type])[disabled]{cursor:not-allowed;background-color:#eaeded;color:#cad2d3}
.pure-form input[readonly],.pure-form select[readonly],.pure-form textarea[readonly]{
background-color:#eee;color:#777;border-color:#ccc}.pure-form input:focus:invalid,
.pure-form textarea:focus:invalid,.pure-form select:focus:invalid{color:#b94a48;border-color:#e9322d}
.pure-form input[type=file]:focus:invalid:focus,.pure-form input[type=radio]:focus:invalid:focus,
.pure-form input[type=checkbox]:focus:invalid:focus{outline-color:#e9322d}.pure-form select{height:2.25em;
border:1px solid #ccc;background-color:#fff}.pure-form select[multiple]{height:auto}.pure-form label{
margin:.5em 0 .2em}.pure-form fieldset{margin:0;padding:.35em 0 .75em;border:0}.pure-form legend{
display:block;width:100%;padding:.3em 0;margin-bottom:.3em;color:#333;border-bottom:1px solid #e5e5e5}
.pure-form-stacked input[type=text],.pure-form-stacked input[type=password],
.pure-form-stacked input[type=email],.pure-form-stacked input[type=url],.pure-form-stacked input[type=date],
.pure-form-stacked input[type=month],.pure-form-stacked input[type=time],
.pure-form-stacked input[type=datetime],.pure-form-stacked input[type=datetime-local],
.pure-form-stacked input[type=week],.pure-form-stacked input[type=number],
.pure-form-stacked input[type=search],.pure-form-stacked input[type=tel],.pure-form-stacked input[type=color],
.pure-form-stacked input[type=file],.pure-form-stacked select,.pure-form-stacked label,
.pure-form-stacked textarea{display:block;margin:.25em 0}.pure-form-stacked input:not([type]){
display:block;margin:.25em 0}.pure-form-aligned input,.pure-form-aligned textarea,.pure-form-aligned select,
.pure-form-aligned .pure-help-inline,.pure-form-message-inline{display:inline-block;*display:inline;*zoom:1;
vertical-align:middle}.pure-form-aligned textarea{vertical-align:top}.pure-form-aligned .pure-control-group{
margin-bottom:.5em}.pure-form-aligned .pure-control-group label{text-align:right;display:inline-block;
vertical-align:middle;width:10em;margin:0 1em 0 0}.pure-form-aligned .pure-controls{margin:1.5em 0 0 11em}
.pure-form input.pure-input-rounded,.pure-form .pure-input-rounded{border-radius:2em;padding:.5em 1em}
.pure-form .pure-group fieldset{margin-bottom:10px}.pure-form .pure-group input,.pure-form .pure-group 
textarea{display:block;padding:10px;margin:0 0 -1px;border-radius:0;position:relative;top:-1px}.pure-form 
.pure-group input:focus,.pure-form .pure-group textarea:focus{z-index:3}.pure-form .pure-group 
input:first-child,.pure-form .pure-group textarea:first-child{top:1px;border-radius:4px 4px 0 0;margin:0}
.pure-form .pure-group input:first-child:last-child,.pure-form .pure-group textarea:first-child:last-child{
top:1px;border-radius:4px;margin:0}.pure-form .pure-group input:last-child,.pure-form .pure-group 
textarea:last-child{top:-2px;border-radius:0 0 4px 4px;margin:0}.pure-form .pure-group button{margin:.35em 0}
.pure-form .pure-input-1{width:100%}.pure-form .pure-input-2-3{width:66%}.pure-form .pure-input-1-2{width:50%}
.pure-form .pure-input-1-3{width:33%}.pure-form .pure-input-1-4{width:25%}.pure-form .pure-help-inline,
.pure-form-message-inline{display:inline-block;padding-left:.3em;color:#666;vertical-align:middle;
font-size:.875em}.pure-form-message{display:block;color:#666;font-size:.875em}
@media only screen and (max-width :480px){.pure-form button[type=submit]{margin:.7em 0 0}
.pure-form input:not([type]),.pure-form input[type=text],.pure-form input[type=password],
.pure-form input[type=email],.pure-form input[type=url],.pure-form input[type=date],
.pure-form input[type=month],.pure-form input[type=time],.pure-form input[type=datetime],
.pure-form input[type=datetime-local],.pure-form input[type=week],.pure-form input[type=number],
.pure-form input[type=search],.pure-form input[type=tel],.pure-form input[type=color],
.pure-form label{margin-bottom:.3em;display:block}.pure-group input:not([type]),.pure-group input[type=text],
.pure-group input[type=password],.pure-group input[type=email],.pure-group input[type=url],
.pure-group input[type=date],.pure-group input[type=month],.pure-group input[type=time],
.pure-group input[type=datetime],.pure-group input[type=datetime-local],.pure-group input[type=week],
.pure-group input[type=number],.pure-group input[type=search],.pure-group input[type=tel],
.pure-group input[type=color]{margin-bottom:0}.pure-form-aligned .pure-control-group label{
margin-bottom:.3em;text-align:left;display:block;width:100%}.pure-form-aligned .pure-controls{
margin:1.5em 0 0}.pure-form .pure-help-inline,.pure-form-message-inline,.pure-form-message{
display:block;font-size:.75em;padding:.2em 0 .8em}}.pure-menu{-webkit-box-sizing:border-box;
-moz-box-sizing:border-box;box-sizing:border-box}.pure-menu-fixed{position:fixed;left:0;top:0;z-index:3}
.pure-menu-list,.pure-menu-item{position:relative}.pure-menu-list{list-style:none;margin:0;padding:0}
.pure-menu-item{padding:0;margin:0;height:100%}.pure-menu-link,.pure-menu-heading{display:block;
text-decoration:none;white-space:nowrap}.pure-menu-horizontal{width:100%;white-space:nowrap}
.pure-menu-horizontal .pure-menu-list{display:inline-block}.pure-menu-horizontal .pure-menu-item,
.pure-menu-horizontal .pure-menu-heading,.pure-menu-horizontal .pure-menu-separator{display:inline-block;
*display:inline;zoom:1;vertical-align:middle}.pure-menu-item .pure-menu-item{display:block}.pure-menu-children{
display:none;position:absolute;left:100%;top:0;margin:0;padding:0;z-index:3}.pure-menu-horizontal
.pure-menu-children{left:0;top:auto;width:inherit}.pure-menu-allow-hover:hover>.pure-menu-children,
.pure-menu-active>.pure-menu-children{display:block;position:absolute}.pure-menu-has-children>
.pure-menu-link:after{padding-left:.5em;content:"\25B8";font-size:small}.pure-menu-horizontal 
.pure-menu-has-children>.pure-menu-link:after{content:"\25BE"}.pure-menu-scrollable{overflow-y:scroll;
overflow-x:hidden}.pure-menu-scrollable .pure-menu-list{display:block}.pure-menu-horizontal
.pure-menu-scrollable .pure-menu-list{display:inline-block}.pure-menu-horizontal.pure-menu-scrollable{
white-space:nowrap;overflow-y:hidden;overflow-x:auto;-ms-overflow-style:none;
-webkit-overflow-scrolling:touch;padding:.5em 0}.pure-menu-horizontal.pure-menu-scrollable::-webkit-scrollbar{
display:none}.pure-menu-separator{background-color:#ccc;height:1px;margin:.3em 0}.pure-menu-horizontal 
.pure-menu-separator{width:1px;height:1.3em;margin:0 .3em}.pure-menu-heading{text-transform:uppercase;
color:#565d64}.pure-menu-link{color:#777}.pure-menu-children{background-color:#fff}.pure-menu-link,
.pure-menu-disabled,.pure-menu-heading{padding:.5em 1em}.pure-menu-disabled{opacity:.5}.pure-menu-disabled 
.pure-menu-link:hover{background-color:transparent}.pure-menu-active>.pure-menu-link,.pure-menu-link:hover,
.pure-menu-link:focus{background-color:#eee}.pure-menu-selected .pure-menu-link,.pure-menu-selected 
.pure-menu-link:visited{color:#000}.pure-table{border-collapse:collapse;border-spacing:0;empty-cells:show;
border:1px solid #cbcbcb}.pure-table caption{color:#000;font:italic 85%/1 arial,sans-serif;padding:1em 0;
text-align:center}.pure-table td,.pure-table th{border-left:1px solid #cbcbcb;border-width:0 0 0 1px;
font-size:inherit;margin:0;overflow:visible;padding:.5em 1em}.pure-table td:first-child,
.pure-table th:first-child{border-left-width:0}.pure-table thead{background-color:#e0e0e0;color:#000;
text-align:left;vertical-align:bottom}.pure-table td{background-color:transparent}.pure-table-odd td{
background-color:#f2f2f2}.pure-table-striped tr:nth-child(2n-1) td{background-color:#f2f2f2}
.pure-table-bordered td{border-bottom:1px solid #cbcbcb}.pure-table-bordered tbody>tr:last-child>td{
border-bottom-width:0}.pure-table-horizontal td,.pure-table-horizontal th{border-width:0 0 1px;
border-bottom:1px solid #cbcbcb}.pure-table-horizontal tbody>tr:last-child>td{border-bottom-width:0}

@media screen and (min-width:35.5em){.pure-u-sm-1,.pure-u-sm-1-1,.pure-u-sm-1-2,.pure-u-sm-1-3,
.pure-u-sm-2-3,.pure-u-sm-1-4,.pure-u-sm-3-4,.pure-u-sm-1-5,.pure-u-sm-2-5,.pure-u-sm-3-5,
.pure-u-sm-4-5,.pure-u-sm-5-5,.pure-u-sm-1-6,.pure-u-sm-5-6,.pure-u-sm-1-8,.pure-u-sm-3-8,
.pure-u-sm-5-8,.pure-u-sm-7-8,.pure-u-sm-1-12,.pure-u-sm-5-12,.pure-u-sm-7-12,.pure-u-sm-11-12,
.pure-u-sm-1-24,.pure-u-sm-2-24,.pure-u-sm-3-24,.pure-u-sm-4-24,.pure-u-sm-5-24,.pure-u-sm-6-24,
.pure-u-sm-7-24,.pure-u-sm-8-24,.pure-u-sm-9-24,.pure-u-sm-10-24,.pure-u-sm-11-24,.pure-u-sm-12-24,
.pure-u-sm-13-24,.pure-u-sm-14-24,.pure-u-sm-15-24,.pure-u-sm-16-24,.pure-u-sm-17-24,.pure-u-sm-18-24,
.pure-u-sm-19-24,.pure-u-sm-20-24,.pure-u-sm-21-24,.pure-u-sm-22-24,.pure-u-sm-23-24,.pure-u-sm-24-24{
display:inline-block;*display:inline;zoom:1;letter-spacing:normal;word-spacing:normal;vertical-align:top;
text-rendering:auto}.pure-u-sm-1-24{width:4.1667%;*width:4.1357%}.pure-u-sm-1-12,.pure-u-sm-2-24{
width:8.3333%;*width:8.3023%}.pure-u-sm-1-8,.pure-u-sm-3-24{width:12.5%;*width:12.469%}.pure-u-sm-1-6,
.pure-u-sm-4-24{width:16.6667%;*width:16.6357%}.pure-u-sm-1-5{width:20%;*width:19.969%}.pure-u-sm-5-24{
width:20.8333%;*width:20.8023%}.pure-u-sm-1-4,.pure-u-sm-6-24{width:25%;*width:24.969%}.pure-u-sm-7-24{
width:29.1667%;*width:29.1357%}.pure-u-sm-1-3,.pure-u-sm-8-24{width:33.3333%;*width:33.3023%}
.pure-u-sm-3-8,.pure-u-sm-9-24{width:37.5%;*width:37.469%}.pure-u-sm-2-5{width:40%;*width:39.969%}
.pure-u-sm-5-12,.pure-u-sm-10-24{width:41.6667%;*width:41.6357%}.pure-u-sm-11-24{width:45.8333%;
*width:45.8023%}.pure-u-sm-1-2,.pure-u-sm-12-24{width:50%;*width:49.969%}.pure-u-sm-13-24{width:54.1667%;
*width:54.1357%}.pure-u-sm-7-12,.pure-u-sm-14-24{width:58.3333%;*width:58.3023%}.pure-u-sm-3-5{
width:60%;*width:59.969%}.pure-u-sm-5-8,.pure-u-sm-15-24{width:62.5%;*width:62.469%}.pure-u-sm-2-3,
.pure-u-sm-16-24{width:66.6667%;*width:66.6357%}.pure-u-sm-17-24{width:70.8333%;*width:70.8023%}
.pure-u-sm-3-4,.pure-u-sm-18-24{width:75%;*width:74.969%}.pure-u-sm-19-24{width:79.1667%;*width:79.1357%}
.pure-u-sm-4-5{width:80%;*width:79.969%}.pure-u-sm-5-6,.pure-u-sm-20-24{width:83.3333%;*width:83.3023%}
.pure-u-sm-7-8,.pure-u-sm-21-24{width:87.5%;*width:87.469%}.pure-u-sm-11-12,.pure-u-sm-22-24{
width:91.6667%;*width:91.6357%}.pure-u-sm-23-24{width:95.8333%;*width:95.8023%}.pure-u-sm-1,.pure-u-sm-1-1,
.pure-u-sm-5-5,.pure-u-sm-24-24{width:100%}}@media screen and (min-width:48em){.pure-u-md-1,.pure-u-md-1-1,
.pure-u-md-1-2,.pure-u-md-1-3,.pure-u-md-2-3,.pure-u-md-1-4,.pure-u-md-3-4,.pure-u-md-1-5,.pure-u-md-2-5,
.pure-u-md-3-5,.pure-u-md-4-5,.pure-u-md-5-5,.pure-u-md-1-6,.pure-u-md-5-6,.pure-u-md-1-8,.pure-u-md-3-8,
.pure-u-md-5-8,.pure-u-md-7-8,.pure-u-md-1-12,.pure-u-md-5-12,.pure-u-md-7-12,.pure-u-md-11-12,
.pure-u-md-1-24,.pure-u-md-2-24,.pure-u-md-3-24,.pure-u-md-4-24,.pure-u-md-5-24,.pure-u-md-6-24,
.pure-u-md-7-24,.pure-u-md-8-24,.pure-u-md-9-24,.pure-u-md-10-24,.pure-u-md-11-24,.pure-u-md-12-24,
.pure-u-md-13-24,.pure-u-md-14-24,.pure-u-md-15-24,.pure-u-md-16-24,.pure-u-md-17-24,.pure-u-md-18-24,
.pure-u-md-19-24,.pure-u-md-20-24,.pure-u-md-21-24,.pure-u-md-22-24,.pure-u-md-23-24,.pure-u-md-24-24{
display:inline-block;*display:inline;zoom:1;letter-spacing:normal;word-spacing:normal;vertical-align:top;
text-rendering:auto}.pure-u-md-1-24{width:4.1667%;*width:4.1357%}.pure-u-md-1-12,.pure-u-md-2-24{
width:8.3333%;*width:8.3023%}.pure-u-md-1-8,.pure-u-md-3-24{width:12.5%;*width:12.469%}.pure-u-md-1-6,
.pure-u-md-4-24{width:16.6667%;*width:16.6357%}.pure-u-md-1-5{width:20%;*width:19.969%}.pure-u-md-5-24{
width:20.8333%;*width:20.8023%}.pure-u-md-1-4,.pure-u-md-6-24{width:25%;*width:24.969%}.pure-u-md-7-24{
width:29.1667%;*width:29.1357%}.pure-u-md-1-3,.pure-u-md-8-24{width:33.3333%;*width:33.3023%}
.pure-u-md-3-8,.pure-u-md-9-24{width:37.5%;*width:37.469%}.pure-u-md-2-5{width:40%;*width:39.969%}
.pure-u-md-5-12,.pure-u-md-10-24{width:41.6667%;*width:41.6357%}.pure-u-md-11-24{width:45.8333%;
*width:45.8023%}.pure-u-md-1-2,.pure-u-md-12-24{width:50%;*width:49.969%}.pure-u-md-13-24{width:54.1667%;
*width:54.1357%}.pure-u-md-7-12,.pure-u-md-14-24{width:58.3333%;*width:58.3023%}.pure-u-md-3-5{width:60%;
*width:59.969%}.pure-u-md-5-8,.pure-u-md-15-24{width:62.5%;*width:62.469%}.pure-u-md-2-3,.pure-u-md-16-24{
width:66.6667%;*width:66.6357%}.pure-u-md-17-24{width:70.8333%;*width:70.8023%}.pure-u-md-3-4,
.pure-u-md-18-24{width:75%;*width:74.969%}.pure-u-md-19-24{width:79.1667%;*width:79.1357%}.pure-u-md-4-5{
width:80%;*width:79.969%}.pure-u-md-5-6,.pure-u-md-20-24{width:83.3333%;*width:83.3023%}.pure-u-md-7-8,
.pure-u-md-21-24{width:87.5%;*width:87.469%}.pure-u-md-11-12,.pure-u-md-22-24{width:91.6667%;
*width:91.6357%}.pure-u-md-23-24{width:95.8333%;*width:95.8023%}.pure-u-md-1,.pure-u-md-1-1,.pure-u-md-5-5,
.pure-u-md-24-24{width:100%}}@media screen and (min-width:64em){.pure-u-lg-1,.pure-u-lg-1-1,.pure-u-lg-1-2,
.pure-u-lg-1-3,.pure-u-lg-2-3,.pure-u-lg-1-4,.pure-u-lg-3-4,.pure-u-lg-1-5,.pure-u-lg-2-5,.pure-u-lg-3-5,
.pure-u-lg-4-5,.pure-u-lg-5-5,.pure-u-lg-1-6,.pure-u-lg-5-6,.pure-u-lg-1-8,.pure-u-lg-3-8,.pure-u-lg-5-8,
.pure-u-lg-7-8,.pure-u-lg-1-12,.pure-u-lg-5-12,.pure-u-lg-7-12,.pure-u-lg-11-12,.pure-u-lg-1-24,
.pure-u-lg-2-24,.pure-u-lg-3-24,.pure-u-lg-4-24,.pure-u-lg-5-24,.pure-u-lg-6-24,.pure-u-lg-7-24,
.pure-u-lg-8-24,.pure-u-lg-9-24,.pure-u-lg-10-24,.pure-u-lg-11-24,.pure-u-lg-12-24,.pure-u-lg-13-24,
.pure-u-lg-14-24,.pure-u-lg-15-24,.pure-u-lg-16-24,.pure-u-lg-17-24,.pure-u-lg-18-24,.pure-u-lg-19-24,
.pure-u-lg-20-24,.pure-u-lg-21-24,.pure-u-lg-22-24,.pure-u-lg-23-24,.pure-u-lg-24-24{display:inline-block;
*display:inline;zoom:1;letter-spacing:normal;word-spacing:normal;vertical-align:top;text-rendering:auto}
.pure-u-lg-1-24{width:4.1667%;*width:4.1357%}.pure-u-lg-1-12,.pure-u-lg-2-24{width:8.3333%;*width:8.3023%}
.pure-u-lg-1-8,.pure-u-lg-3-24{width:12.5%;*width:12.469%}.pure-u-lg-1-6,.pure-u-lg-4-24{width:16.6667%;
*width:16.6357%}.pure-u-lg-1-5{width:20%;*width:19.969%}.pure-u-lg-5-24{width:20.8333%;*width:20.8023%}
.pure-u-lg-1-4,.pure-u-lg-6-24{width:25%;*width:24.969%}.pure-u-lg-7-24{width:29.1667%;*width:29.1357%}
.pure-u-lg-1-3,.pure-u-lg-8-24{width:33.3333%;*width:33.3023%}.pure-u-lg-3-8,.pure-u-lg-9-24{width:37.5%;
*width:37.469%}.pure-u-lg-2-5{width:40%;*width:39.969%}.pure-u-lg-5-12,.pure-u-lg-10-24{width:41.6667%;
*width:41.6357%}.pure-u-lg-11-24{width:45.8333%;*width:45.8023%}.pure-u-lg-1-2,.pure-u-lg-12-24{width:50%;
*width:49.969%}.pure-u-lg-13-24{width:54.1667%;*width:54.1357%}.pure-u-lg-7-12,.pure-u-lg-14-24{
width:58.3333%;*width:58.3023%}.pure-u-lg-3-5{width:60%;*width:59.969%}.pure-u-lg-5-8,.pure-u-lg-15-24{
width:62.5%;*width:62.469%}.pure-u-lg-2-3,.pure-u-lg-16-24{width:66.6667%;*width:66.6357%}.pure-u-lg-17-24{
width:70.8333%;*width:70.8023%}.pure-u-lg-3-4,.pure-u-lg-18-24{width:75%;*width:74.969%}.pure-u-lg-19-24{
width:79.1667%;*width:79.1357%}.pure-u-lg-4-5{width:80%;*width:79.969%}.pure-u-lg-5-6,.pure-u-lg-20-24{
width:83.3333%;*width:83.3023%}.pure-u-lg-7-8,.pure-u-lg-21-24{width:87.5%;*width:87.469%}.pure-u-lg-11-12,
.pure-u-lg-22-24{width:91.6667%;*width:91.6357%}.pure-u-lg-23-24{width:95.8333%;*width:95.8023%}
.pure-u-lg-1,.pure-u-lg-1-1,.pure-u-lg-5-5,.pure-u-lg-24-24{width:100%}}@media screen and (min-width:80em){
.pure-u-xl-1,.pure-u-xl-1-1,.pure-u-xl-1-2,.pure-u-xl-1-3,.pure-u-xl-2-3,.pure-u-xl-1-4,.pure-u-xl-3-4,
.pure-u-xl-1-5,.pure-u-xl-2-5,.pure-u-xl-3-5,.pure-u-xl-4-5,.pure-u-xl-5-5,.pure-u-xl-1-6,.pure-u-xl-5-6,
.pure-u-xl-1-8,.pure-u-xl-3-8,.pure-u-xl-5-8,.pure-u-xl-7-8,.pure-u-xl-1-12,.pure-u-xl-5-12,
.pure-u-xl-7-12,.pure-u-xl-11-12,.pure-u-xl-1-24,.pure-u-xl-2-24,.pure-u-xl-3-24,.pure-u-xl-4-24,
.pure-u-xl-5-24,.pure-u-xl-6-24,.pure-u-xl-7-24,.pure-u-xl-8-24,.pure-u-xl-9-24,.pure-u-xl-10-24,
.pure-u-xl-11-24,.pure-u-xl-12-24,.pure-u-xl-13-24,.pure-u-xl-14-24,.pure-u-xl-15-24,.pure-u-xl-16-24,
.pure-u-xl-17-24,.pure-u-xl-18-24,.pure-u-xl-19-24,.pure-u-xl-20-24,.pure-u-xl-21-24,.pure-u-xl-22-24,
.pure-u-xl-23-24,.pure-u-xl-24-24{display:inline-block;*display:inline;zoom:1;letter-spacing:normal;
word-spacing:normal;vertical-align:top;text-rendering:auto}.pure-u-xl-1-24{width:4.1667%;*width:4.1357%}
.pure-u-xl-1-12,.pure-u-xl-2-24{width:8.3333%;*width:8.3023%}.pure-u-xl-1-8,.pure-u-xl-3-24{width:12.5%;
*width:12.469%}.pure-u-xl-1-6,.pure-u-xl-4-24{width:16.6667%;*width:16.6357%}.pure-u-xl-1-5{width:20%;
*width:19.969%}.pure-u-xl-5-24{width:20.8333%;*width:20.8023%}.pure-u-xl-1-4,.pure-u-xl-6-24{width:25%;
*width:24.969%}.pure-u-xl-7-24{width:29.1667%;*width:29.1357%}.pure-u-xl-1-3,.pure-u-xl-8-24{
width:33.3333%;*width:33.3023%}.pure-u-xl-3-8,.pure-u-xl-9-24{width:37.5%;*width:37.469%}.pure-u-xl-2-5{
width:40%;*width:39.969%}.pure-u-xl-5-12,.pure-u-xl-10-24{width:41.6667%;*width:41.6357%}.pure-u-xl-11-24{
width:45.8333%;*width:45.8023%}.pure-u-xl-1-2,.pure-u-xl-12-24{width:50%;*width:49.969%}.pure-u-xl-13-24{
width:54.1667%;*width:54.1357%}.pure-u-xl-7-12,.pure-u-xl-14-24{width:58.3333%;*width:58.3023%}
.pure-u-xl-3-5{width:60%;*width:59.969%}.pure-u-xl-5-8,.pure-u-xl-15-24{width:62.5%;*width:62.469%}
.pure-u-xl-2-3,.pure-u-xl-16-24{width:66.6667%;*width:66.6357%}.pure-u-xl-17-24{width:70.8333%;
*width:70.8023%}.pure-u-xl-3-4,.pure-u-xl-18-24{width:75%;*width:74.969%}.pure-u-xl-19-24{width:79.1667%;
*width:79.1357%}.pure-u-xl-4-5{width:80%;*width:79.969%}.pure-u-xl-5-6,.pure-u-xl-20-24{width:83.3333%;
*width:83.3023%}.pure-u-xl-7-8,.pure-u-xl-21-24{width:87.5%;*width:87.469%}.pure-u-xl-11-12,
.pure-u-xl-22-24{width:91.6667%;*width:91.6357%}.pure-u-xl-23-24{width:95.8333%;*width:95.8023%}
.pure-u-xl-1,.pure-u-xl-1-1,.pure-u-xl-5-5,.pure-u-xl-24-24{width:100%}}
`
}
