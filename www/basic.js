jQuery(function($) {

let salt, popo_offset = 20;

let ic = window.ic = {
  get,
  ajaxform,
  init,
  password,
  loadpage,
  smartTable,
  deleteButton,
  popo,
  ynDialog,
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
      if (data.code) {
        cb(new Error('错误:'+ data.msg));
      } else {
        cb(null, data)
      }
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

  jdom.find(":input").focus(function() {
    msg.text('');
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
// <table api='获取数据的接口' form='查询条件表单选择器(可选)'
//        pageinfo='获取分页数据接口'>
// <thead><th cn='来自数据接口中的列属性名' 
//         rename='重命名属性' format='显示时格式化' >...
// format: date 
//
// table 事件:
//    error(event, errObj)     : 解析数据异常, 运行异常
//    refresh_success          : 数据刷新成功
//    select_row(event, rowData, tr_jdom): 行被选中
//
// table 绑定属性:
//    refreshData() 立即刷新数据
//
function smartTable(jdom) {
  const api  = jdom.attr('api');
  let form   = $(jdom.attr('form') || '<form>');
  let tbody  = jdom.find("tbody");
  let header = [];
  let rows   = 0;
  let oldtr  = jdom.find("tr");
  let totalpage = 0;
  let currpage = 0;
  let pagedom;
  let errmsg;
  let search_filter_changed;

  jdom.refreshData = refresh_data;

  let no_format = function(d) { return d; };
  let format_fn = {
    'date': function(d) {
      let t = new Date(d);
      if (isNaN(t.getDate())) return d;
      return t.toLocaleString();
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

  let pageinfo = jdom.attr('pageinfo');
  if (pageinfo) {
    get(pageinfo, {}, function(err, ret) {
      if (err) return _error(err);
      if (ret.code) return _error(new Error(ret.msg));
      totalpage = Math.ceil(ret.data.Count / ret.data.PageSize) || 0;
      init_page();
    });
  }

  form.submit(function() {
    if (search_filter_changed) {
      search_filter_changed = false;
      currpage = 0;
    }
    refresh_data();
    return false;
  });

  form.find(":input").change(function() {
    search_filter_changed = true;
  });

  setTimeout(refresh_data, 1);

  function _error(err) {
    if (errmsg) {
      errmsg.text(err.message);
    } else {
      tbody.append("<tr><td>"+ err.message +"</td></tr>");
    }
    jdom.trigger('error', err);
  }

  function refresh_data() {
    let param = form.serialize();
    param += "&page=" + currpage;

    get(api, param, function(err, ret) {
      update_page();
      if (err) return _error(err);
      tbody.html('');
      if (!ret.data) return _error(new Error("没有数据"));
      
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

  function init_page() {
    pagedom = get_template("#page_template");
    errmsg = pagedom.find(".error-message");
    jdom.after(pagedom);
    update_page();

    pagedom.find('.prev').click(function() {
      currpage--;
      refresh_data();
    });

    pagedom.find('.next').click(function() {
      currpage++;
      refresh_data();
    });

    pagedom.submit(function() {
      let ipage = pagedom.find(".currpage");
      let pn = ipage.val()-1;
      if (pn != currpage && pn >= 0 && pn < totalpage) {
        currpage = pn;
        refresh_data();
      } else {
        ipage.val(currpage+1);
      }
      return false;
    });
  }

  function update_page() {
    if (!pagedom) return;
    if (currpage < 0) currpage = 0;
    else if (currpage >= totalpage) currpage = totalpage-1;
    pagedom.find(".currpage").val(currpage + 1);
    pagedom.find('.prev').prop("disabled", currpage-1 < 0);
    pagedom.find('.next').prop("disabled", currpage+1 >= totalpage);
    errmsg.text('');
  }
}


function get_template(selector) {
  let t = $(selector).clone();
  t.removeClass('html_template');
  t.removeAttr("id");
  return t;
}

//
// <button api='删除接口' ...
// jdel 必须绑定数据 .data('id', ...) 该参数直接递交到接口
// 事件:
//  delete_success(event) : 删除成功
//  error(event, err)     : 删除时发生错误
//
function deleteButton(jdel) {
  const api = jdel.attr('api');

  jdel.click(function() {
    let id = jdel.data("id");
    if (!id) return popo(new Error(".data('id', ... 绑定参数无效"));
    const title = ["删除选中的数据? 数据删除后将不可还原",
      "<br/>选择 '是' 删除 <b style='color:#d04242'>", 
      id, "</b>"].join('');

    ynDialog(title, function(err, yn) {
        if (!yn) return;
        let parm = {'id': id};
        get(api, parm, function(err, ret) {
          if (err) {
            jdel.trigger("error", err);
            return popo(err);
          }
          jdel.trigger("delete_success");
          popo(id +" 数据被删除");
        });
      });
  });
}


//
// 弹出消息气泡, msg: 字符串/错误对象
//
function popo(msg) {
  if (!msg) return;
  let t = get_template('#popo_message_template');
  let conf;
  let height = 0;
  let bd = $(document.body);
  let content = t.find('.popo_message');
  let mouseon;
  let offset = popo_offset;

  bd.append(t);
  bd.on("popo_message_removed", on_other_removed);

  setTimeout(()=>{
    content.addClass("show");
    content.css("bottom", (offset) +'px');
    height = content.outerHeight() + 20;
    popo_offset += height;
  }, 10);

  let tid = setInterval(()=>{
    if (mouseon) return;
    clearInterval(tid);
    bd.off("popo_message_removed", on_other_removed);

    content.stop().animate({'bottom': '-100px'}, 300, function() {
      t.remove();
      popo_offset -= height;
      bd.trigger("popo_message_removed", height);
    });
  }, 3e3);

  content.mouseenter(()=>{
    mouseon = 1;
  });

  content.mouseleave(()=>{
    mouseon = 0;
  });

  if (msg.constructor == Error) {
    conf = ['错误', msg.message, 'error'];
  } else if (typeof msg == "string") {
    conf = ['消息', msg, 'info'];
  } else {
    conf = ['调试', JSON.stringify(msg), 'debug'];
  }

  t.find(".ti").text(conf[0]);
  t.find(".msg").html(conf[1]);
  content.addClass(conf[2]);

  function on_other_removed(ev, mheight) {
    offset -= mheight;
    content.animate({'bottom': (offset) +'px'}, 1e3);
  }
}


//
// 一个 是/否 选择框
// cb: Function(err, yes_no_bool)
//
function ynDialog(msg, cb) {
  let t = get_template('#select_yn_dialog');
  $(document.body).append(t);
  t.find('.message').html(msg);

  setTimeout(() => {
    t.find('.select_yn_dialog').addClass("move");
  }, 10);

  t.find(".yes").click(function() {
    t.hide(100, function() { t.remove() });
    cb(null, true);
  });

  t.find(".no").click(function() {
    t.hide(100, function() { t.remove() });
    cb(null, false);
  });
}

});