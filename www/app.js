jQuery(function($) {


ic.init(function(err) {
  if (err) {
    alert(err.message);
  } else {
    checkHash();
  }
});
  
ic.get('whoaim', null, function(err, d) {
  if (err || d.code) {
    alert("用户未登录, 即将跳转到登录页面");
    location.href = 'index.html';
  } else {
    ic.username = d.data;
  }
});

$('#main_menu .s2').click(function() {
  let thiz = $(this);
  let href = thiz.attr("href");
  if (!href) return;
  ic.loadpage(href);
  location.hash = href;
});

function checkHash() {
  let h = location.hash;
  if (h[0] == '#') h = h.substr(1);
  if (h) {
    ic.loadpage(h);
  }
}

});