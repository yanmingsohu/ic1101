jQuery(function($) {

const API_ROOT = "/ic/";
const body = $(document.body);
const win  = $(window);
const content_frame = $("#main_frame");

const ic = window.ic = {
  get,
  getDict,
  ajaxform,
  init,
  password,
  loadpage,
  smartTable,
  deleteButton,
  popo,
  ynDialog,
  contentDialog,
  commandCrudPage,
  getTemplate: get_template,
  buildSelectOpt,
  initDictSelect,
  select2fromApi,
  numberScope,
  dateTime,
};

const format_fn = {
  'date': function(d) {
    let t = new Date(d);
    if (isNaN(t.getDate())) return d;
    return t.toLocaleString();
  },
  'yesno': function(d) {
    return Boolean(d) ? '是':'否';
  },
  'duration': function(d) {
    d /= 1e6;
    if (d < 1000) return d.toFixed(2) +"毫秒";
    d /= 1e3;
    if (d < 60) return d.toFixed(2) +"秒";
    d /= 60;
    if (d < 60) return d.toFixed(2) +"分钟";
    d /= 60;
    if (d < 24) return d.toFixed(2) +"小时";
    d /= 24;
    return d.toFixed(2) +"天";
  },
};


function init(cb) {
  ic.get('salt', null, function(err, ret) {
    if (err) return cb(new Error("应用初始化失败"));
    body.data('salt', ret.data);
    cb(null, ret);
    // console.log(salt)
  });
  body.data('popo_offset', 20);
}


function password(name, pass) {
  var spark = new SparkMD5();
  spark.append(name);
  spark.append(pass);
  spark.append(body.data('salt'));
  return spark.end();
}


function get(api, data, cb) {
  $.ajax(API_ROOT + api, {
    type : 'GET',
    data : data,

    success : function(data, status, jxr) {
      if (data.code === 0) {
        cb(null, data)
      } else {
        let type = jxr.getResponseHeader("content-type");
        if (type && (type.indexOf("json") < 0)) {
          cb(new Error("错误的应答, 不是 json. fail:"+ type));
          return;
        }

        let err = new Error('错误:'+ data.msg);
        err.data = data.data;
        cb(err);
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
    if (msg.length) {
      msg.text(err.message);
    } else {
      popo(err);
    }
    jdom.trigger("error", err);
  }
}


//
// <table api='获取数据的接口' form='查询条件表单选择器(可选)'
//        pageinfo='获取分页数据接口'>
// <thead><th cn='来自数据接口中的列属性名' 
//         rename='重命名属性' format='显示时格式化' >...
// format: date / yesno / 在 data("format.XXX", function ColConv) 绑定方法
//    Function ColConv(value) return other-value
//
// table 事件:
//    error(event, errObj)     : 解析数据异常, 运行异常
//    refresh_success          : 数据刷新成功
//    select_row(event, rowData, tr_jdom): 行被选中, rowData 已经转换
//
// table 绑定属性:
//    refreshData() 立即刷新数据
//    data('select_row') 当前选择行的数据(已经被 _convert 转换)
//    selectNone()  取消当前选择的行
//
// _convert: 可选的数据转换器, 必须返回数组, 默认返回 api 返回的数组
//
function smartTable(jdom, _convert) {
  const api    = jdom.attr('api');
  const form   = $(jdom.attr('form') || '<form>');
  const tbody  = jdom.find("tbody");
  const header = [];
  let rows   = 0;
  let oldtr  = jdom.find("tr");
  let totalpage = 0;
  let currpage = 0;
  let pagedom;
  let search_filter_changed;
  let convert_data = _convert || _get_array_data;

  jdom.refreshData = refresh_data;
  jdom.selectNone = selectNone;

  if (!api) {
    return _error(new Error("缺少 api 属性"));
  }

  let no_format = function(d) { return d; };

  jdom.find("thead th").each(function() {
    let thiz = $(this);
    let format_name = thiz.attr('format');
    let format_func = no_format;

    if (format_name) {
      if (format_name.startsWith("format.")) {
        let _f = jdom.data(format_name);
        if (typeof _f == 'function') format_func = _f;
      } 
      else if (format_fn[format_name]) {
        format_func = format_fn[format_name];
      }
    }
    
    header.push({
      cn : thiz.attr('cn'),
      rn : thiz.attr('rename'),
      fm : format_func,
    });
  });

  let pageinfo = jdom.attr('pageinfo');
  if (pageinfo) {
    get(pageinfo, {}, function(err, ret) {
      if (err) return _error(err);
      if (ret.code) return _error(new Error(ret.msg));
      totalpage = Math.ceil(ret.data.Count / ret.data.PageSize) || 1;
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
    popo(err);
    jdom.trigger('error', err);
    console.error(jdom, err);
  }

  function _get_array_data(d) {
    return d.data;
  }

  function refresh_data() {
    let param = form.serialize();
    param += "&page=" + currpage;

    get(api, param, function(err, ret) {
      update_page();
      if (err) return _error(err);
      tbody.html('');
      if (!ret.data) return _error(new Error("没有数据"));
      
      convert_data(ret).forEach(function(v, i) {
        let tr = $("<tr>").appendTo(tbody);
        tr.data("rowData", v);
        if (rows++ & 1) {
          tr.addClass('pure-table-odd');
        }
        if (v == null) return;

        header.forEach(function(h) {
          let td = $("<td>").appendTo(tr);
          td.html(h.fm( v[h.cn] ));
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
      jdom.data('select_row', value);
      jdom.trigger('select_row', value, tr);
    });
  }

  function init_page() {
    pagedom = get_template("#page_template");
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
        ipage.val(currpage +1);
      }
      return false;
    });
  }

  function selectNone() {
    oldtr.removeClass("table_row_click");
    oldtr = jdom.find("tr");
    jdom.removeData('select_row');
  }

  function update_page() {
    if (!pagedom) return;
    if (currpage < 0) currpage = 0;
    else if (currpage >= totalpage) currpage = totalpage-1;
    pagedom.find(".currpage").val(currpage + 1);
    pagedom.find('.prev').prop("disabled", currpage-1 < 0);
    pagedom.find('.next').prop("disabled", currpage+1 >= totalpage);
  }
}


//
// 带有 html_template 的模板对象进行复制
//
function get_template(selector) {
  let t = $(selector).clone();
  if (t.length == 0) {
    throw new Error("错误: HTML 模板标签不存在, "+ selector);
  }
  t.removeClass('html_template');
  t.removeAttr("id");
  return t;
}

//
// <button api='删除接口' ...
// jdel 必须绑定数据 .data('id', ...) 该参数直接递交到接口
// .data('parm', ...) 递交的扩展参数
// .data('what', ...) 提示用户要删除什么的消息字符串, 默认提示为 id
// 事件:
//  delete_success(event) : 删除成功
//  error(event, err)     : 删除时发生错误
//  cancel(event)         : 取消删除
//
function deleteButton(jdel) {
  const api = jdel.attr('api');

  jdel.click(function() {
    let id = jdel.data("id");
    if (!id) return popo(new Error(".data('id', ... 绑定参数无效"));
    let what = jdel.data("what") || id;

    const title = ["删除选中的数据? 数据删除后将不可还原",
      "<br/>选择 '是' 删除 <b style='color:#d04242'>", 
      what, "</b>"].join('');

    ynDialog(title, function(err, yn) {
        if (!yn) return jdel.trigger('cancel');
        let parm = $.extend({'id': id}, jdel.data('parm'));
        
        get(api, parm, function(err, ret) {
          if (err) {
            jdel.trigger("error", err);
            return popo(err);
          }
          jdel.trigger("delete_success");
          if (ret.msg) {
            let txt = ret.msg;
            if (ret.data) txt += "<br/>"+ ret.data;
            popo(txt);
          } else {
            popo(id +" 数据被删除");
          }
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
  let content = t.find('.popo_message');
  let mouseon;
  let offset = body.data('popo_offset');

  // content_frame.append(t); // 子页面会遮挡弹出消息
  body.append(t);
  body.on("popo_message_removed", on_other_removed);

  setTimeout(_show, 10);
  t.click(_close);

  let tid = setInterval(()=>{
    if (mouseon) return;
    _close();
  }, 5e3);

  content.mouseenter(()=>{
    mouseon = 1;
  });

  content.mouseleave(()=>{
    mouseon = 0;
  });

  if (msg.constructor == Error) {
    conf = ['错误', msg.message, 'error'];
    if (msg.data) {
      let buf = ["<div class='small'>"];
      if (typeof msg.data == 'string') {
        buf.push(msg.data);
      } else {
        for (let n in msg.data) {
          buf.push(n, ": ", msg.data[n], ", ");
        }
      }
      buf.push("</div>");
      conf[1] += buf.join("");
    }
  } else if (typeof msg == "string") {
    conf = ['消息', msg, 'info'];
  } else {
    conf = ['调试', JSON.stringify(msg), 'debug'];
  }

  t.find(".ti").text(conf[0]);
  t.find(".msg").html(conf[1]);
  content.addClass(conf[2]);

  function _show() {
    content.addClass("show");
    content.css("bottom", (offset) +'px');
    height = content.outerHeight() + 20;
    add_popo_offset(height);
  }

  function _close() {
    clearInterval(tid);
    body.off("popo_message_removed", on_other_removed);

    content.stop().animate({'bottom': '-100px'}, 300, function() {
      t.remove();
      add_popo_offset(-height);
      body.trigger("popo_message_removed", height);
    });
  }

  function on_other_removed(ev, mheight) {
    offset -= mheight;
    content.animate({'bottom': (offset) +'px'}, 1e3);
  }

  function add_popo_offset(x) {
    let v = body.data('popo_offset');
    body.data('popo_offset', v + x);
  }
}


//
// 一个 是/否 选择框
// cb: Function(err, yes_no_bool)
//
function ynDialog(msg, cb) {
  let t = get_template('#select_yn_dialog');
  content_frame.append(t);
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


//
// 创建一个对话框, 内容从 url 加载, 返回显示对话框内容的框架 jquery 对象
// 事件, 通过 on() 接受:
//   opend(event) : 窗口正确打开
//   error(event, error) : 发生错误
//   closing(event) : 窗口正在关闭, 此时抛出异常可终止关闭
//   closed(event) : 窗口已经关闭
//
// 方法, 通过 trigger 调用:
//   close : 立即关闭窗口
//
function contentDialog(url) {
  if (!url) throw new Error("url not null");

  let t = get_template("#content_dialog_template");
  content_frame.append(t);
  let content = t.find(".content_dialog"); 
  let loading = t.find(".loading");
  let close = t.find(".closeframe *");
  resize();
  win.resize(resize);
  content.animate({right: 0}, 200);


  $.ajax(url, {
    type : 'GET',
    success : _open_page,
    error : function(req, text, errText) {
      let e = new Error(text +", "+ errText);
      loading.hide();
      content.find('.content').html(url +"<br/>"+ text +"<br/>"+ errText);
      t.trigger('error', e);
    },
  });

  close.click(function() {
    try {
      t.trigger('closing');
      content.animate({right: -1000}, 200);
      t.fadeOut(500, function() {
        t.trigger('closed');
        win.off('resize', resize);
        t.remove();
      });
    } catch(err) {}
    return false;
  });

  function resize() {
    content.width(content_frame.outerWidth());
    content.height(window.innerHeight);
  }

  function _open_page(html) {
    let c = content.find('.content');
    try {
      loading.hide();
      c.html(html);
      t.trigger('opend');
    } catch(err) {
      c.html('<pre>'+ err.stack +"</pre>");
      t.trigger('error', err);
    }
  }
  return t;
}


//
// 把 who 居中到 where, 带有 5像素 下移动画
//
function center(where, who) {
  let a = [where.height(), where.width()];
  let b = [who.outerHeight(), who.outerWidth()];
  let off = { 
    top : (a[0]-b[0])/2 - 5, 
    left: (a[1]-b[1])/2 
  };
  who.offset(off);

  setTimeout(function() {
    off.top += 5;
    who.animate(off, 100, 'swing');
  }, 1);
}


//
// 通常增删改查页面初始化
// conf: {
//   create: 该对象处理方法由 create_page 属性决定, 表单/按钮
//   create_page: (可选的) 如果有该属性则 create 是按钮, 并在按下后打开页面
//   edit  : 编辑按钮对象
//   edit_page : 编辑子页面
//   delete : 删除按钮对象
//   table : 表格对象
//   table_convert : 表格数据转换器
//   copy_edit: function(data, targetDom) : 
//      用当前数据初始化编辑子页面, 适用于 create_page/edit_page
// }
//
// table 附加事件: button_disabled(event, bool)
//
function commandCrudPage(conf) {
  smartTable(conf.table, conf.table_convert);
  deleteButton(conf.delete);
  update_button(true);

  conf.delete.on('delete_success', refreshTableData);

  if (conf.create_page) {
    conf.create.click(function() {
      openDialogWith(conf.create_page);
    });
  } else {
    ajaxform(conf.create);
    conf.create.on('success', refreshTableData);
  }

  conf.table.on('select_row', function(_, v) {
    update_button((v || v.id) == null);
    conf.delete.data('id', v && v.id);
  });

  conf.table.on('refresh_success', function() {
    update_button(true);
  });

  conf.edit.click(function() {
    openDialogWith(conf.edit_page);
  });

  function openDialogWith(page) {
    let dialog = contentDialog(page);
    dialog.on('opend', function() {
      let data = conf.table.data('select_row');
      conf.copy_edit(data, dialog);
    });
    dialog.on('closed', refreshTableData);
  }

  function refreshTableData() {
    conf.table.refreshData();
  }

  function update_button(dis) {
    conf.edit.prop('disabled', dis);
    conf.delete.prop('disabled', dis);
    conf.table.trigger("button_disabled", dis);
  }
}


//
// 查字典, cb: Function(error, dictMap)
//
function getDict(dictId, cb) {
  if (!dictId) throw new Error("dict ID is null");

  const key = "ic1101.dict."+ dictId;
  let cache = JSON.parse(sessionStorage.getItem(key));
  if (cache) {
    cb(null, cache);
    return;
  }

  get('dict_read', {id: dictId}, function(err, ret) {
    if (err) return cb(err);
    let d = ret.data;
    sessionStorage.setItem(key, JSON.stringify(d));
    cb(null, d);
  });
}


//
// 选择字典的列表
//
function initDictSelect(jselect) {
  jselect.attr("api", "dict_list");
  select2fromApi(jselect, function(r) {
    if (r.data) {
      r.data.forEach(function(d) {
        let txt = [d._id];
        if (d.desc) {
          txt.push(' - ', d.desc);
        }
        d.id = d._id;
        d.text = txt.join("");
      });
      return r.data;
    }
  });
}


//
// <select api='' 绑定到 select2 下拉菜单, 数据来自 api, 分页.
// data_convert: Function(json_data) 转换json数据使之有 {id, text}
//              返回 null 则没有更多数据
//
// 接口应接受 page 参数用于翻页, text 用于全文检索
// 用 trigger('change', value) 触发默认选项的加载, 
// 一旦数据加载完成, 触发 change_over(event, data) 事件
//
function select2fromApi(jselect, data_convert) {
  const api = jselect.attr('api');
  if (!api) return popo("api 参数无效");

  if (!data_convert) {
    // 默认转换器, 取返回数据中的 ret.data, 
    // 并且绑定 id/text, id=_id, text = _id + desc
    data_convert = function(ret) {
      if (ret && ret.data && ret.data.forEach) {
        ret.data.forEach(function(d) {
          let txt = d._id;
          if (d.desc) txt += ' - '+ d.desc;
          d.id = d._id;
          d.text = txt;
        });
        // console.log("data_convert", ret);
        return ret.data;
      }
    };
  }

  jselect.select2({
    ajax: { 
      url: API_ROOT + api,
      dataType: 'json',
      quietMillis: 250,

      data: function (t) {
        return {
          text : t.term,
          page : t.page ? t.page-1 : 0,
        };
      },

      processResults: function (r, q) {
        let ret = { results: [], pagination: {more: false} };
        let dc = data_convert(r);
        if (dc) {
          ret.pagination.more = true;
          ret.results = dc;
        }
        return ret;
      },
    },
    cache: true,
  });

  if (jselect.val()) {
    set_value(jselect.val());
  }

  jselect.on("change", function(event, value) {
    if (!value) return;
    set_value(value);
  });

  function set_value(value) {
    $.get(API_ROOT + api, {text: value}, function(ret) {
      let d = data_convert(ret);
      let row = d && d[0];
      if (!row) return;

      jselect.html('');
      var opt = new Option(row.text, row.id, true, true);
      jselect.append(opt);
      jselect.trigger("change_over", row);
    });
  }
}


//
// map[k] = v, k 作为 value, v 作为显示
//
function buildSelectOpt(jselect, map, value) {
  for (let k in map) {
    let opt = $("<option>");
    opt.val(k);
    opt.text(map[k]);
    jselect.append(opt);
  }
  if (value) jselect.val(value).trigger("change");
  return jselect;
}


function numberScope(bit, signed) {
  let r = {min:0, max:0};

  if (signed) {
    if (bit == 32) {
      r.min = "-2147483648";
      r.max = "2147483647";
    } else if (bit == 64) {
      r.min = "-9223372036854775808";
      r.max = "9223372036854775807";
    } else {
      let x = (1 << bit)>>1;
      r.max = x-1;
      r.min = -x;
    }
  } else {
    if (bit == 64) {
      r.max = "18446744073709552000";
    } else if (bit == 32) {
      r.max = 0xffffffff;
    } else {
      r.max = (1 << bit)-1;
    }
  }
  return r;
}


//
// 文本控件转换为时间控件, 正确处理时区, 
// 递交格式为: RFC1123 / GMT
//
function dateTime(jinput) {
  let dt = $("<input type='datetime-local'>");
  jinput.hide().after(dt);
  dt.change(syncVal);

  let date = new Date(jinput.val());
  if (!date.getDate()) {
    date = new Date();
  }

  let buf = [ date.getFullYear(), '-', s2(date.getMonth()+1), '-', 
              s2(date.getDate()), 'T', s2(date.getHours()), ':', 
              s2(date.getMinutes()) ];
  dt.val(buf.join(""));
  copyAttr("class");
  copyAttr("placeholder");
  syncVal();

  function copyAttr(name) {
    dt.attr(name, jinput.attr(name));
  }

  function syncVal() {
    let d = new Date(dt.val());
    jinput.val(d.toGMTString());
  }

  function s2(x) {
    if (x < 10) return '0'+ x;
    return ''+ x;
  }
}

});