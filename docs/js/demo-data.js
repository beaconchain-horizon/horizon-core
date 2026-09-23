// داده‌های نمایشی برای حالت دمو
window.DEMO = {
  tps: 847,
  blocks: 1247,
  sensors: 24,
  readings: 128493,
  recentBlocks: [
    {num:1247, hash:'0x6eda20e975f80dec0794af183fc732ad', tx:12, time:'۲ دقیقه پیش'},
    {num:1246, hash:'0x9b1976ddb9c279ee257c828f9ea6de30', tx:8,  time:'۵ دقیقه پیش'},
    {num:1245, hash:'0x8247b1838140ccbeb907fd712f9c3a11', tx:15, time:'۸ دقیقه پیش'},
    {num:1244, hash:'0xc85838c1beb5602bf960275b41c5e9a2', tx:6,  time:'۱۲ دقیقه پیش'},
    {num:1243, hash:'0xabcf9d1e7f4c3a5b2e8d1f6c9b7a3e5d', tx:22, time:'۱۵ دقیقه پیش'}
  ],
  recentTx: [
    {from:'بانک مرکزی', to:'بانک ملی', amount:1000000, type:'transfer'},
    {from:'بانک ملی', to:'بانک صادرات', amount:500000, type:'transfer'},
    {from:'پالایش و پخش', to:'بانک مرکزی', amount:2500000, type:'settlement'},
    {from:'ملی گاز', to:'بانک ملت', amount:750000, type:'transfer'}
  ],
  sensors: [
    {id:'temp-001', site:'پالایشگاه امام خمینی', type:'دما', value:67.3, unit:'°C', status:'ok', min:20, max:80},
    {id:'press-002', site:'پالایشگاه امام خمینی', type:'فشار', value:8.4, unit:'bar', status:'ok', min:1, max:10},
    {id:'flow-003', site:'ملی گاز', type:'جریان', value:1245, unit:'m³/h', status:'warn', min:500, max:2000},
    {id:'vib-004', site:'فولاد مبارکه', type:'لرزش', value:0.45, unit:'mm/s', status:'ok', min:0, max:2},
    {id:'gas-005', site:'پتروشیمی', type:'گاز', value:0.02, unit:'ppm', status:'critical', min:0, max:0.01},
    {id:'temp-006', site:'صنایع مس', type:'دما', value:812, unit:'°C', status:'ok', min:700, max:900}
  ],
  alerts: [
    {sev:'critical', sensor:'gas-005', msg:'نشت گاز · مقدار 0.02 ppm', time:'۲ دقیقه پیش'},
    {sev:'warning', sensor:'flow-003', msg:'نزدیک به مرز بالا · 1245 m³/h', time:'۸ دقیقه پیش'},
    {sev:'info', sensor:'temp-001', msg:'پایداری دما · 67.3°C', time:'۱۵ دقیقه پیش'}
  ],
  industries: [
    {name:'بانک مرکزی', type:'bank', icon:'fa-landmark'},
    {name:'بانک ملی', type:'bank', icon:'fa-university'},
    {name:'پالایش و پخش', type:'oil', icon:'fa-oil-well'},
    {name:'ملی گاز', type:'gas', icon:'fa-fire'},
    {name:'فولاد مبارکه', type:'steel', icon:'fa-hard-hat'},
    {name:'صنایع مس', type:'copper', icon:'fa-gem'}
  ]
};

// تابع برای شبیه‌سازی API
window.fetchDemoData = function(endpoint) {
  return new Promise(resolve => {
    setTimeout(() => {
      if (endpoint.includes('stats')) resolve({chainLength: DEMO.blocks, tps: DEMO.tps, recentBlocks: DEMO.recentBlocks});
      else if (endpoint.includes('sensors')) resolve(DEMO.sensors);
      else if (endpoint.includes('alerts')) resolve(DEMO.alerts);
      else resolve(DEMO);
    }, 300);
  });
};
