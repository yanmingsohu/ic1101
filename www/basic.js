jQuery(function($) {

let salt;

let ic = window.ic = {
  get,
  ajaxform,
  init,
  password,
  loadpage,
};


function init(cb) {
  ic.get('salt', null, function(err, ret) {
    if (err) return cb(new Error("应用初始化失败"));
    salt = ret.data;
    cb(null, ret);
    // console.log(salt)
  });
}


function password(name, pass) {
  var spark = new SparkMD5();
  spark.append(name);
  spark.append(pass);
  spark.append(salt);
  return spark.end();
}


function get(api, data, cb) {
  $.ajax("/ic/"+ api, {
    type : 'GET',
    data : data,

    success : function(data) {
      cb(null, data)
    },

    error : function(req, text, err) {
      cb(err || (new Error(text)))
    },
  });
}


function loadpage(page, cb) {
  let main_frame = $("#main_frame");
  main_frame.html("");

  $.get(page, function(data) {
    main_frame.append(data);
    cb && cb();
  }, 'html');
}


//
// 事件, 在调用该方法前绑定:
//  before : (event)           在递交前触发
//  param : (event, paramData) 用于在递交前修改参数
//  error : (event, errorObj)  递交时发生 http 错误
//  success : (event, data)    递交后返回数据(完整的 json 应答)
//
function ajaxform(jdom) {
  let api = jdom.attr('action');
  let msg = jdom.find('.ajax-message');

  msg.text("");

  jdom.submit(function() {
    try {
      jdom.trigger('before');
    } catch(err) {
      gotError(err);
    }
    return false;
  });

  jdom.on('before', function() {
    let param = jdom.serialize();
    jdom.trigger("param", param);
  });

  jdom.on('param', function(event, param) {
    get(api, param, function(err, ret) {
      if (err) {
        gotError(err);
        return;
      }

      setErrorFlag(ret.code);
      msg.text(ret.msg);
      jdom.trigger('success', ret);
    });
  });

  function setErrorFlag(is) {
    if (is) {
      msg.addClass('error-message');
    } else {
      msg.removeClass('error-message');
    }
  }

  function gotError(err) {
    setErrorFlag(true);
    msg.text(err.message);
    jdom.trigger("error", err);
  }
}

});