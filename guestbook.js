'use strict';
var btn = document.getElementById("btn");
var guestbook = document.getElementById("guestbook");
var serverURL = ""

function myCallback(acptlang) {
  if (acptlang.Message != "") {
    alert(acptlang.Message);
  }
  console.log(acptlang);
  guestbook.innerHTML = "";
  console.log(acptlang.Entries.length);
  for (var i = 0; i < acptlang.Entries.length; i++) {
    guestbook.innerHTML = guestbook.innerHTML + ` <li class="pa3 pa4-ns bb b--black-10"> <span class="f5 db lh-copy measure i">${acptlang.Entries[i].Message}</span> <span class="f5 db lh-copy measure tr"> - ${acptlang.Entries[i].Name}</span> <span class="f5 db lh-copy measure tr">${acptlang.Entries[i].DateString}</span> <span class="f5 db lh-copy measure tr">${acptlang.Entries[i].Location}</span></li>`

  }
}

function jsonp() {
  guestbook.innerHTML = "Loading ...";
  var tag = document.createElement("script");
  var message = encodeURIComponent(document.querySelector('#message').value);
  var name = encodeURIComponent(document.querySelector('#name').value);
  var email = encodeURIComponent(document.querySelector('#email').value);
  tag.src = `${serverURL}/jsonp?callback=myCallback&message=${message}&name=${name}&email=${name}`;
  document.querySelector("head").appendChild(tag);
}
btn.addEventListener("click", jsonp);
window.onload = jsonp;
