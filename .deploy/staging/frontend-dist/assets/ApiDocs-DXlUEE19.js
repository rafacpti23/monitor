import{_ as v,c as o,a as e,F as c,m as r,e as i,t,h as y,o as n,s as m,d as l,r as b}from"./index-Cuz32HcK.js";const T={class:"apidocs"},E={class:"doc-layout"},g={class:"doc-nav"},A=["href","onClick"],k={class:"doc-content"},P={class:"doc-header"},w=["id"],S={key:0,class:"group-desc"},C={class:"ep-head"},G={class:"ep-path"},O={class:"ep-desc"},R={key:0,class:"ep-body"},L={__name:"ApiDocs",setup(f){const u="https://monitor.papi.api.br/api/v1",p=b("auth"),h=[{id:"auth",title:"Autenticação",desc:"Registro, login e sessão do usuário.",endpoints:[{method:"POST",path:"/auth/register",desc:"Cria uma empresa e um usuário admin.",body:`{
  "email": "you@empresa.com",
  "name": "Seu Nome",
  "password": "min 8 chars",
  "company_name": "Acme Corp"
}`},{method:"POST",path:"/auth/login",desc:"Retorna token JWT + dados do usuário + empresa (whitelabel).",body:`{
  "email": "you@empresa.com",
  "password": "..."
}`},{method:"GET",path:"/auth/me",desc:"Retorna o usuário atual e a empresa. Requer token."},{method:"POST",path:"/auth/logout",desc:"Encerra a sessão."}]},{id:"company",title:"Empresa (Whitelabel)",desc:"Marca personalizada e gestão de usuários internos. Escrita exige perfil admin.",endpoints:[{method:"GET",path:"/company",desc:"Retorna a empresa (nome, system_name, logo_url, accent_color)."},{method:"PUT",path:"/company",desc:"Atualiza marca. Admin.",body:`{
  "name": "Acme Corp",
  "system_name": "Acme Monitor",
  "logo_url": "https://.../logo.png",
  "accent_color": "#00e676"
}`},{method:"GET",path:"/company/users",desc:"Lista usuários internos da empresa."},{method:"POST",path:"/company/users",desc:"Cria usuário interno. Admin.",body:`{
  "email": "user@empresa.com",
  "name": "Nome",
  "password": "min 8",
  "role": "member",
  "whatsapp_number": "5511999999999"
}`},{method:"PUT",path:"/company/users/:id",desc:"Edita usuário. Senha em branco = mantém. Admin.",body:`{
  "name": "Novo Nome",
  "role": "viewer",
  "whatsapp_number": "...",
  "password": ""
}`},{method:"DELETE",path:"/company/users/:id",desc:"Remove usuário (não pode ser você mesmo). Admin."}]},{id:"servers",title:"Servidores",endpoints:[{method:"GET",path:"/servers",desc:"Lista servidores com métricas mais recentes."},{method:"POST",path:"/servers",desc:"Cria servidor e gera server_key.",body:`{
  "name": "vps-prod-01",
  "type": "linux"
}`},{method:"GET",path:"/servers/:id",desc:"Detalhe do servidor + serviços."},{method:"PUT",path:"/servers/:id",desc:"Atualiza servidor."},{method:"DELETE",path:"/servers/:id",desc:"Remove servidor."},{method:"GET",path:"/servers/:id/history",desc:"Histórico de métricas. Query: ?range=1h|6h|24h|7d|30d"},{method:"GET",path:"/servers/:id/incidents",desc:"Incidentes do servidor."}]},{id:"agent",title:"Agente (Receiver)",desc:"Endpoint público usado pelo binário agente instalado no VPS. Aceita JSON ou JSON gzipado.",endpoints:[{method:"POST",path:"/agent/:key",desc:"Recebe payload de métricas. A key é o server_key do servidor.",body:`{
  "hostname": "vps-prod-01",
  "os": "Ubuntu 24.04",
  "cpu_percent": 34.7,
  "ram_total_bytes": 1073741824,
  "ram_used_bytes": 536870912,
  "load_avg": [0.42, 0.38, 0.35],
  "disks": [{"path":"/","total_bytes":..,"used_bytes":..}],
  "docker_containers": [...],
  "pm2_processes": [...],
  "services": [{"name":"nginx","running":true}]
}`}]},{id:"websites",title:"Websites",endpoints:[{method:"GET",path:"/websites",desc:"Lista websites monitorados."},{method:"POST",path:"/websites",desc:"Adiciona website.",body:`{
  "name": "Meu Site",
  "url": "https://exemplo.com",
  "method": "GET",
  "check_interval_sec": 60,
  "search_string": ""
}`},{method:"GET",path:"/websites/:id",desc:"Detalhe do website."},{method:"PUT",path:"/websites/:id",desc:"Atualiza website."},{method:"DELETE",path:"/websites/:id",desc:"Remove website."},{method:"GET",path:"/websites/:id/history",desc:"Histórico de respostas."}]},{id:"checks",title:"Checks",desc:"Verificações de rede: ping, tcp, http, dns, ssl_expiry.",endpoints:[{method:"GET",path:"/checks",desc:"Lista checks."},{method:"POST",path:"/checks",desc:"Cria check.",body:`{
  "name": "DNS Google",
  "type": "dns",
  "target": "google.com",
  "port": 0,
  "interval_sec": 60
}`},{method:"GET",path:"/checks/:id",desc:"Detalhe do check."},{method:"PUT",path:"/checks/:id",desc:"Atualiza check."},{method:"DELETE",path:"/checks/:id",desc:"Remove check."}]},{id:"incidents",title:"Incidentes",endpoints:[{method:"GET",path:"/incidents",desc:"Lista incidentes."},{method:"PUT",path:"/incidents/:id/acknowledge",desc:"Reconhece incidente."},{method:"PUT",path:"/incidents/:id/resolve",desc:"Resolve incidente."},{method:"PUT",path:"/incidents/:id/ignore",desc:"Ignora incidente."}]},{id:"alerts",title:"Regras & Canais de Alerta",desc:"Canais suportados: email, whatsapp (PAPI), webhook.",endpoints:[{method:"GET",path:"/alert-rules",desc:"Lista regras de alerta."},{method:"POST",path:"/alert-rules",desc:"Cria regra.",body:`{
  "monitor_type": "server",
  "monitor_id": 1,
  "alert_type": "cpu",
  "comparison": ">=",
  "threshold": "90",
  "occurrences": 3,
  "cooldown_min": 30,
  "channels": "[1]"
}`},{method:"PUT",path:"/alert-rules/:id",desc:"Atualiza regra."},{method:"DELETE",path:"/alert-rules/:id",desc:"Remove regra."},{method:"GET",path:"/settings/channels",desc:"Lista canais de notificação."},{method:"POST",path:"/settings/channels",desc:"Cria canal. Para WhatsApp use type=whatsapp e config PAPI.",body:`// WhatsApp (PAPI)
{
  "type": "whatsapp",
  "name": "Alertas WhatsApp",
  "config": "{\\"instance\\":\\"minha_instancia\\",\\"api_key\\":\\"SUA_API_KEY\\",\\"jid\\":\\"5511999999999\\"}",
  "enabled": true
}`},{method:"PUT",path:"/settings/channels/:id",desc:"Atualiza canal."},{method:"DELETE",path:"/settings/channels/:id",desc:"Remove canal."},{method:"POST",path:"/settings/channels/:id/test",desc:"Envia mensagem de teste pelo canal."}]},{id:"papi",title:"WhatsApp via PAPI",desc:"O único provedor de WhatsApp suportado. Você configura apenas: Instância, API Key e JID (número que recebe o alerta).",endpoints:[{method:"POST",path:"https://api.papi.api.br/api/instances/{instancia}/send-text",desc:"Chamada real feita pelo backend ao disparar um alerta. Header x-api-key, corpo jid + text.",body:`curl -X POST "https://api.papi.api.br/api/instances/{instancia}/send-text" \\
  -H "x-api-key: SUA_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "jid": "5511999999999",
    "text": "Olá do Papi!"
  }'`}]}];return(x,d)=>(n(),o("div",T,[e("div",E,[e("nav",g,[(n(),o(c,null,r(h,s=>e("a",{key:s.id,href:"#"+s.id,class:m(["doc-nav-item",{active:p.value===s.id}]),onClick:a=>p.value=s.id},t(s.title),11,A)),64))]),e("div",k,[e("header",P,[d[1]||(d[1]=e("h1",null,"Documentação da API",-1)),e("p",null,[d[0]||(d[0]=i("Base URL: ",-1)),e("code",null,t(y(u)),1)]),d[2]||(d[2]=e("p",{class:"auth-note"},[i(" Autenticação via "),e("code",null,"Authorization: Bearer <token>"),i(" (obtido no login). O agente usa a rota pública "),e("code",null,"/agent/:key"),i(" com a chave do servidor. ")],-1))]),(n(),o(c,null,r(h,s=>e("section",{key:s.id,id:s.id,class:"doc-section"},[e("h2",null,t(s.title),1),s.desc?(n(),o("p",S,t(s.desc),1)):l("",!0),(n(!0),o(c,null,r(s.endpoints,(a,_)=>(n(),o("div",{key:_,class:"endpoint"},[e("div",C,[e("span",{class:m(["method",a.method.toLowerCase()])},t(a.method),3),e("code",G,t(a.path),1)]),e("p",O,t(a.desc),1),a.body?(n(),o("pre",R,[e("code",null,t(a.body),1)])):l("",!0)]))),128))],8,w)),64))])])]))}},U=v(L,[["__scopeId","data-v-61519564"]]);export{U as default};
