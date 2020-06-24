jQuery(function($) {

let salt;

let ic = window.ic = {
  get,
  ajaxform,
  init,
  password,
  loadpage,
  smartTable,
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


//
// <table api='获取数据的接口' form='查询条件表单选择器(可选)'>
// <thead><th cn='来自数据接口中的列属性名' 
//         rename='重命名属性' format='显示时格式化' >...
// format: date 
//
// table 事件:
//   error(event, errObj)     : 解析数据异常, 运行异常
//   refresh_success          : 数据刷新成功
//   select_row(event, rowData, tr_jdom): 行被选中
//
function smartTable(jdom) {
  let api = jdom.attr('api');
  let form = $(jdom.attr('form'));
  let tbody = jdom.find("tbody");
  let header = [];
  let param = { page: 0 };
  let rows = 0;
  let oldtr = jdom.find("tr");

  let no_format = function(d) { return d; };
  let format_fn = {
    'data': function(d) {
      return new Date(d).toLocaleString();
    },
  };

  jdom.find("thead th").each(function() {
    let thiz = $(this);
    header.push({
      cn : thiz.attr('cn'),
      rn : thiz.attr('rename'),
      fm : format_fn[thiz.attr('format')] || no_format,
    });
  });

  setTimeout(refresh_data, 1);

  function refresh_data() {
    get(api, param, function(err, ret) {
      if (err) {
        tbody.append("<tr><td>"+ err.message +"</td></tr>");
        jdom.trigger('error', err);
        return;
      }
      
      ret.data.forEach(function(v, i) {
        let tr = $("<tr>").appendTo(tbody);
        tr.data("rowData", v);
        if (rows++ & 1) {
          tr.addClass('pure-table-odd');
        }

        header.forEach(function(h) {
          let td = $("<td>").appendTo(tr);
          td.text(h.fm( v[h.cn] ));
          if (h.rn) {
            v[h.rn] = v[h.cn];
          }
        });

        row_click(tr, v);
      });
      jdom.trigger('refresh_success');
    });
  }

  function row_click(tr, value) {
    tr.click(function() {
      oldtr.removeClass("table_row_click");
      tr.addClass("table_row_click");
      oldtr = tr;
      jdom.trigger('select_row', value, tr);
    });
  }

  return {
    jdom,
    refresh_data,
  };
}

});