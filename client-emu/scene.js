(function() {
//
// 佳木斯学校供暖
//
function on_data(dev, time, data) {
  switch (dev.GetName()) {
    
    case '回水压力':
    case '供水压力':
      data /= 100; // 缩小100倍
      break;
      
    case '阀门反馈开度':
    case '手动设定开度':
      
    case '总表一次供温':
    case '总表一次回温':
      
    case '总表累计流量':
    case '总表累计热量':
      data /= 10; // 缩小10倍
      break;
  }
  return data;
}

})()