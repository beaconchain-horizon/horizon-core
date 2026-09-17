package main

const industrialPanelHTML = `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Horizon Industrial Panel</title>
<style>
  *{box-sizing:border-box}
  body{font-family:-apple-system,Tahoma,sans-serif;background:#0f172a;color:#e2e8f0;margin:0;padding:24px}
  h1{color:#38bdf8;margin:0 0 24px}
  h2{color:#94a3b8;font-size:14px;text-transform:uppercase;letter-spacing:1px;margin:24px 0 12px}
  .grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:12px}
  .card{background:#1e293b;border-radius:10px;padding:14px;border:1px solid #334155}
  .card h3{margin:0 0 6px;color:#38bdf8;font-size:15px}
  .card .meta{color:#94a3b8;font-size:12px}
  .alert{padding:10px 14px;border-radius:8px;margin:6px 0;font-size:13px}
  .critical{background:#7f1d1d;border-right:4px solid #ef4444}
  .warning{background:#78350f;border-right:4px solid #f59e0b}
  .info{background:#1e3a5f;border-right:4px solid #3b82f6}
  .stat{display:inline-block;margin-left:24px}
  .stat b{color:#38bdf8;font-size:22px;display:block}
  .muted{color:#64748b;font-size:12px}
</style>
</head>
<body>
<h1>Horizon Industrial Panel</h1>
<div id="stats"></div>
<h2>Sites</h2>
<div class="grid" id="sites"></div>
<h2>Sensors</h2>
<div class="grid" id="sensors"></div>
<h2>Recent Alerts</h2>
<div id="alerts"></div>
<p class="muted">Auto-refresh every 5s</p>
<script>
function esc(s){return String(s==null?'':s).replace(/[&<>"']/g,function(c){return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c];});}
function renderStats(d){
  document.getElementById('stats').innerHTML=
    '<div class="stat"><b>'+(d.sites?d.sites.length:0)+'</b>Sites</div>'+
    '<div class="stat"><b>'+(d.sensors?d.sensors.length:0)+'</b>Sensors</div>'+
    '<div class="stat"><b>'+(d.alerts?d.alerts.length:0)+'</b>Alerts</div>'+
    '<div class="stat"><b>'+(d.readings_count||0)+'</b>Readings</div>'+
    '<div class="stat"><b>'+(d.tamper_count||0)+'</b>Tamper Events</div>';
}
function renderSites(a){
  var el=document.getElementById('sites');
  if(!a.length){el.innerHTML='<div class="card">No sites</div>';return;}
  el.innerHTML=a.map(function(s){return '<div class="card"><h3>'+esc(s.name)+'</h3><div class="meta">'+esc(s.site_id)+' — '+esc(s.type)+'</div></div>';}).join('');
}
function renderSensors(a){
  var el=document.getElementById('sensors');
  if(!a.length){el.innerHTML='<div class="card">No sensors</div>';return;}
  el.innerHTML=a.map(function(s){return '<div class="card"><h3>'+esc(s.name)+'</h3><div class="meta">'+esc(s.sensor_id)+'</div><div class="meta">'+esc(s.type)+' / '+esc(s.unit)+' ['+s.min_value+'..'+s.max_value+']</div></div>';}).join('');
}
function renderAlerts(a){
  var el=document.getElementById('alerts');
  if(!a.length){el.innerHTML='<div class="card">No alerts</div>';return;}
  el.innerHTML=a.slice(0,25).map(function(x){
    var c=x.severity==='critical'?'critical':(x.severity==='warning'?'warning':'info');
    return '<div class="alert '+c+'"><b>'+esc(x.sensor_id)+'</b> — '+esc(x.message)+'</div>';
  }).join('');
}
async function load(){
  try{
    var r=await fetch("/api/v1/industrial/dashboard");
    if(r.status===401){document.getElementById("stats").innerHTML="<div class="muted">dashboard محافظت‌شده است (نیاز به توکن ادمین)</div>";return;}
    var d=await r.json();
    renderStats(d);renderSites(d.sites||[]);renderSensors(d.sensors||[]);renderAlerts(d.alerts||[]);
  }catch(e){console.error(e);}
}
load();setInterval(load,5000);
</script>
</body>
</html>`
