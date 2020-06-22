jQuery(function($) {


let ic = window.ic = {
  get,
  ajaxform,
};


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


//
// 事件, 在调用该方法前绑定:
//  before : (event)            在递交前触发
//  param : (event, paramData) 用于在递交前修改参数
//  error : (event, errorObj)  递交时发生 http 错误
//  success : (event, data)    递交后返回数据(完整的 json 应答)
//
function ajaxform(jdom) {
  let api = jdom.attr('action');
  let msg = jdom.find('.ajax-message');

  msg.text("");

  jdom.submit(function() {
    jdom.trigger('before');
    return false;
  });

  jdom.on('before', function() {
    let param = jdom.serialize();
    jdom.trigger("param", param);
  });

  jdom.on('param', function(event, param) {
    get(api, param, function(err, ret) {
      if (err) {
        msg.text(err.message);
        jdom.trigger("error", err);
        return;
      }

      if (ret.code) {
        msg.addClass('error-message');
      } else {
        msg.removeClass('error-message');
      }

      msg.text(ret.msg);
      jdom.trigger('success', ret);
    });
  });
}

});