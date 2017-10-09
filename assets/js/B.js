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
