'use strict';
var btn = document.getElementById("btn");
var guestbook = document.getElementById("guestbook");
var serverURL = "http://localhost:8054"

function myCallback(acptlang) {
  if (acptlang.Message != "") {
    alert(acptlang.Message);
  }
  console.log(acptlang);
  guestbook.innerHTML = "";
  console.log(acptlang.Entries.length);
  for (var i = 0; i < acptlang.Entries.length; i++) {
    guestbook.innerHTML = guestbook.innerHTML + `<div class="guestbook-comment"> <div class="guestbook-comment-text"> ${acptlang.Entries[i].Message}</div><div class="guestbook-comment-meta"> - ${acptlang.Entries[i].Name}</div><div class="guestbook-comment-meta">${acptlang.Entries[i].DateString}</div><div class="guestbook-comment-meta">${acptlang.Entries[i].Location}</div></div>`

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
