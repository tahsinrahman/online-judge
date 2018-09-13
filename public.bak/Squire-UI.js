$(document).ready(function() {
  Squire.prototype.testPresenceinSelection = function(name, action, format,
    validation) {
    var path = this.getPath(),
      test = (validation.test(path) | this.hasFormat(format));
    if (name == action && test) {
      return true;
    } else {
      return false;
    }
  };
  SquireUI = function(options) {
    if (typeof options.buildPath == "undefined") {
      options.buildPath = 'build/';
    }
    // Create instance of iFrame
    var container, editor;
    if (options.replace) {
      container = $(options.replace).parent();
      $(options.replace).remove();
    } else if (options.div) {
      container = $(options.div);
    } else {
      throw new Error(
        "No element was defined for the editor to inject to.");
    }
    var iframe = document.createElement('iframe');
    var div = document.createElement('div');
    div.className = 'Squire-UI';
    iframe.height = options.height;

    $(div).load(options.buildPath + 'Squire-UI.html', function() {
      this.linkDrop = new Drop({
        target: $('#makeLink').first()[0],
        content: $('#drop-link').html(),
        position: 'bottom center',
        openOn: 'click'
      });

      this.linkDrop.on('open', function () {
        $('.quit').click(function () {
          $(this).parent().parent().removeClass('drop-open');
        });

        $('.submitLink').click(function () {
          var editor = iframe.contentWindow.editor;
          editor.makeLink($(this).parent().children('#url').first().val());
          $(this).parent().parent().removeClass('drop-open');
          $(this).parent().children('#url').attr('value', '');
        });
      });

      this.imageDrop = new Drop({
        target: $('#insertImage').first()[0],
        content: $('#drop-image').html(),
        position: 'bottom center',
        openOn: 'click'
      });

      this.imageDrop.on('open', function () {
        $('.quit').unbind().click(function () {
          $(this).parent().parent().removeClass('drop-open');
        });

        $('.sumbitImageURL').unbind().click(function () {
          console.log("Passed through .sumbitImageURL");
          var editor = iframe.contentWindow.editor;
          url = $(this).parent().children('#imageUrl').first()[0];
          editor.insertImage(url.value);
          $(this).parent().parent().removeClass('drop-open');
          $(this).parent().children('#imageUrl').attr('value', '');
        });

      });

      this.fontDrop = new Drop({
        target: $('#selectFont').first()[0],
        content: $('#drop-font').html(),
        position: 'bottom center',
        openOn: 'click'
      });

      this.fontDrop.on('open', function () {
        $('.quit').click(function () {
          $(this).parent().parent().removeClass('drop-open');
        });

        $('.submitFont').unbind().click(function () {
          var editor = iframe.contentWindow.editor;
          var selectedFonts = $('select#fontSelect option:selected').last().data('fonts');
          var fontSize =  $('select#textSelector option:selected').last().data('size') + 'px';
          editor.setFontSize(fontSize);

          try {
            editor.setFontFace(selectedFonts);
          } catch (e) {
            alert('Please make a selection of text.');
          } finally {
            $(this).parent().parent().removeClass('drop-open');
          }

        });


      });

      $('.item').click(function() {
        var iframe = $(this).parents('.Squire-UI').next('iframe').first()[0];
        var editor = iframe.contentWindow.editor;
        var action = $(this).data('action');

        test = {
          value: $(this).data('action'),
          testBold: editor.testPresenceinSelection('bold',
            action, 'B', (/>B\b/)),
          testItalic: editor.testPresenceinSelection('italic',
            action, 'I', (/>I\b/)),
          testUnderline: editor.testPresenceinSelection(
            'underline', action, 'U', (/>U\b/)),
          testOrderedList: editor.testPresenceinSelection(
            'makeOrderedList', action, 'OL', (/>OL\b/)),
          testLink: editor.testPresenceinSelection('makeLink',
            action, 'A', (/>A\b/)),
          testQuote: editor.testPresenceinSelection(
            'increaseQuoteLevel', action, 'blockquote', (
              />blockquote\b/)),
          isNotValue: function (a) {return (a == action && this.value !== ''); }
        };

        editor.alignRight = function () { editor.setTextAlignment('right'); };
        editor.alignCenter = function () { editor.setTextAlignment('center'); };
        editor.alignLeft = function () { editor.setTextAlignment('left'); };
        editor.alignJustify = function () { editor.setTextAlignment('justify'); };
        editor.makeHeading = function () { editor.setFontSize('2em'); editor.bold(); };

        if (test.testBold | test.testItalic | test.testUnderline | test.testOrderedList | test.testLink | test.testQuote) {
          if (test.testBold) editor.removeBold();
          if (test.testItalic) editor.removeItalic();
          if (test.testUnderline) editor.removeUnderline();
          if (test.testLink) editor.removeLink();
          if (test.testOrderedList) editor.removeList();
          if (test.testQuote) editor.decreaseQuoteLevel();
        } else if (test.isNotValue('makeLink') | test.isNotValue('insertImage') | test.isNotValue('selectFont')) {
          // do nothing these are dropdowns.
        } else {
            editor[action]();
            editor.focus();
        }
      });
    });

    iframe.addEventListener('load', function() {
      // Make sure we're in standards mode.
      var doc = iframe.contentDocument;
      if ( doc.compatMode !== 'CSS1Compat' ) {
          doc.open();
          doc.write( '<!DOCTYPE html><title></title>' );
          doc.close();
      }
      // doc.close() can cause a re-entrant load event in some browsers,
      // such as IE9.
      if ( iframe.contentWindow.editor ) {
          return;
      }
      iframe.contentWindow.editor = new Squire(iframe.contentWindow.document);
      iframe.contentWindow.editor.addStyles(
          'html {' +
          '  height: 100%;' +
          '}' +
          'body {' +
          '  -moz-box-sizing: border-box;' +
          '  -webkit-box-sizing: border-box;' +
          '  box-sizing: border-box;' +
          '  height: 100%;' +
          '  padding: 1em;' +
          '  background: transparent;' +
          '  color: #2b2b2b;' +
          '  font: 13px/1.35 Helvetica, arial, sans-serif;' +
          '  cursor: text;' +
          '}' +
          'a {' +
          '  text-decoration: underline;' +
          '}' +
          'h1 {' +
          '  font-size: 138.5%;' +
          '}' +
          'h2 {' +
          '  font-size: 123.1%;' +
          '}' +
          'h3 {' +
          '  font-size: 108%;' +
          '}' +
          'h1,h2,h3,p {' +
          '  margin: 1em 0;' +
          '}' +
          'h4,h5,h6 {' +
          '  margin: 0;' +
          '}' +
          'ul, ol {' +
          '  margin: 0 1em;' +
          '  padding: 0 1em;' +
          '}' +
          'blockquote {' +
          '  border-left: 2px solid blue;' +
          '  margin: 0;' +
          '  padding: 0 10px;' +
          '}'
      );
    });

    $(container).append(div);
    $(container).append(iframe);

    return iframe.contentWindow.editor;
  };
});
/*! drop 0.5.4 */
!function(t,e){"function"==typeof define&&define.amd?define(e):"object"==typeof exports?module.exports=e(require,exports,module):t.Tether=e()}(this,function(){return function(){var t,e,o,i,n,s,r,l,h,a,p,f,u,d,c,g,m,b={}.hasOwnProperty,v=[].indexOf||function(t){for(var e=0,o=this.length;o>e;e++)if(e in this&&this[e]===t)return e;return-1},y=[].slice;null==this.Tether&&(this.Tether={modules:[]}),p=function(t){var e,o,i,n,s;if(o=getComputedStyle(t).position,"fixed"===o)return t;for(i=void 0,e=t;e=e.parentNode;){try{n=getComputedStyle(e)}catch(r){}if(null==n)return e;if(/(auto|scroll)/.test(n.overflow+n["overflow-y"]+n["overflow-x"])&&("absolute"!==o||"relative"===(s=n.position)||"absolute"===s||"fixed"===s))return e}return document.body},c=function(){var t;return t=0,function(){return t++}}(),m={},h=function(t){var e,i,s,r,l;if(s=t._tetherZeroElement,null==s&&(s=t.createElement("div"),s.setAttribute("data-tether-id",c()),n(s.style,{top:0,left:0,position:"absolute"}),t.body.appendChild(s),t._tetherZeroElement=s),e=s.getAttribute("data-tether-id"),null==m[e]){m[e]={},l=s.getBoundingClientRect();for(i in l)r=l[i],m[e][i]=r;o(function(){return m[e]=void 0})}return m[e]},u=null,r=function(t){var e,o,i,n,s,r,l;t===document?(o=document,t=document.documentElement):o=t.ownerDocument,i=o.documentElement,e={},l=t.getBoundingClientRect();for(n in l)r=l[n],e[n]=r;return s=h(o),e.top-=s.top,e.left-=s.left,null==e.width&&(e.width=document.body.scrollWidth-e.left-e.right),null==e.height&&(e.height=document.body.scrollHeight-e.top-e.bottom),e.top=e.top-i.clientTop,e.left=e.left-i.clientLeft,e.right=o.body.clientWidth-e.width-e.left,e.bottom=o.body.clientHeight-e.height-e.top,e},l=function(t){return t.offsetParent||document.documentElement},a=function(){var t,e,o,i,s;return t=document.createElement("div"),t.style.width="100%",t.style.height="200px",e=document.createElement("div"),n(e.style,{position:"absolute",top:0,left:0,pointerEvents:"none",visibility:"hidden",width:"200px",height:"150px",overflow:"hidden"}),e.appendChild(t),document.body.appendChild(e),i=t.offsetWidth,e.style.overflow="scroll",s=t.offsetWidth,i===s&&(s=e.clientWidth),document.body.removeChild(e),o=i-s,{width:o,height:o}},n=function(t){var e,o,i,n,s,r,l;for(null==t&&(t={}),e=[],Array.prototype.push.apply(e,arguments),l=e.slice(1),s=0,r=l.length;r>s;s++)if(i=l[s])for(o in i)b.call(i,o)&&(n=i[o],t[o]=n);return t},d=function(t,e){var o,i,n,s,r;if(null!=t.classList){for(s=e.split(" "),r=[],i=0,n=s.length;n>i;i++)o=s[i],o.trim()&&r.push(t.classList.remove(o));return r}return t.className=t.className.replace(new RegExp("(^| )"+e.split(" ").join("|")+"( |$)","gi")," ")},e=function(t,e){var o,i,n,s,r;if(null!=t.classList){for(s=e.split(" "),r=[],i=0,n=s.length;n>i;i++)o=s[i],o.trim()&&r.push(t.classList.add(o));return r}return d(t,e),t.className+=" "+e},f=function(t,e){return null!=t.classList?t.classList.contains(e):new RegExp("(^| )"+e+"( |$)","gi").test(t.className)},g=function(t,o,i){var n,s,r,l,h,a;for(s=0,l=i.length;l>s;s++)n=i[s],v.call(o,n)<0&&f(t,n)&&d(t,n);for(a=[],r=0,h=o.length;h>r;r++)n=o[r],a.push(f(t,n)?void 0:e(t,n));return a},i=[],o=function(t){return i.push(t)},s=function(){var t,e;for(e=[];t=i.pop();)e.push(t());return e},t=function(){function t(){}return t.prototype.on=function(t,e,o,i){var n;return null==i&&(i=!1),null==this.bindings&&(this.bindings={}),null==(n=this.bindings)[t]&&(n[t]=[]),this.bindings[t].push({handler:e,ctx:o,once:i})},t.prototype.once=function(t,e,o){return this.on(t,e,o,!0)},t.prototype.off=function(t,e){var o,i,n;if(null!=(null!=(i=this.bindings)?i[t]:void 0)){if(null==e)return delete this.bindings[t];for(o=0,n=[];o<this.bindings[t].length;)n.push(this.bindings[t][o].handler===e?this.bindings[t].splice(o,1):o++);return n}},t.prototype.trigger=function(){var t,e,o,i,n,s,r,l,h;if(o=arguments[0],t=2<=arguments.length?y.call(arguments,1):[],null!=(r=this.bindings)?r[o]:void 0){for(n=0,h=[];n<this.bindings[o].length;)l=this.bindings[o][n],i=l.handler,e=l.ctx,s=l.once,i.apply(null!=e?e:this,t),h.push(s?this.bindings[o].splice(n,1):n++);return
