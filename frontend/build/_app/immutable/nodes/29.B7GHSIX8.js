import{a as d,c as ie,f as _}from"../chunks/C1g-UAMB.js";import{o as de}from"../chunks/WGvgV8zC.js";import{p as le,b as p,e as ce,f as R,a as pe,h as ge,g as a,i as r,s as i,c as n,$ as ue,r as o,o as j,t as M}from"../chunks/zQIQxtAy.js";import{d as ve,a as h,s as U,e as me}from"../chunks/BdS1jnjP.js";import{i as W}from"../chunks/D6vd17M5.js";import{e as be}from"../chunks/gPghANsq.js";import{h as ye}from"../chunks/Bml8Zt1z.js";import{r as q,b as xe}from"../chunks/Dp7jx1s0.js";import{s as fe}from"../chunks/BGje0cyo.js";import{b as he}from"../chunks/0jamiZ8b.js";import{a as K}from"../chunks/nosI4nge.js";import{D as ke}from"../chunks/CZ3-aSlY.js";import{M as _e}from"../chunks/B_241PC_.js";import{t as k}from"../chunks/BakoPrhB.js";var we=_('<button type="button"> </button>'),P=_("<pre><code></code></pre>"),Te=_('<div class="space-y-4"><div class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-4"><p class="text-xs font-semibold text-emerald-600 dark:text-emerald-400">Token Generated Successfully!</p> <p class="mt-1 text-xs text-gray-600 dark:text-gray-300">Copy this token now. For security purposes, you will not be able to view it again.</p> <div class="mt-3 flex items-center gap-2"><input type="text" readonly="" class="flex-1 font-mono text-xs rounded-lg border border-gray-300 bg-gray-50 px-3 py-2 text-emerald-600 focus:outline-none select-all dark:border-gray-700 dark:bg-gray-900 dark:text-emerald-300"/> <button type="button" class="rounded-lg bg-teal-600 px-3 py-2 text-xs font-semibold text-white shadow-sm transition hover:bg-teal-500">Copy</button></div></div> <div class="flex justify-end pt-2"><button type="button" class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700 transition hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200">Done</button></div></div>'),Oe=_('<form class="space-y-4"><div><label for="token-label" class="mb-1.5 block text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">Token Name / Service</label> <input id="token-label" type="text" placeholder="e.g. ERP Backend Gateway, Mobile App" class="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2.5 text-sm text-gray-900 placeholder-gray-400 focus:border-teal-600 focus:outline-none focus:ring-1 focus:ring-teal-600 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"/></div> <div class="flex justify-end gap-3 pt-2"><button type="button" class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700 transition hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700">Cancel</button> <button type="submit" class="rounded-lg bg-teal-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-teal-500 disabled:opacity-50"> </button></div></form>'),Pe=_('<div class="space-y-8"><div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div><h1 class="text-xl font-bold tracking-tight text-gray-950 dark:text-white">Developer API Access</h1> <p class="text-sm text-gray-500 dark:text-gray-400">Manage high-speed Bearer authentication tokens for automated SMS dispatch.</p></div> <button type="button" class="flex items-center gap-2 rounded-lg bg-teal-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-teal-500"><svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"></path></svg> Generate New Token</button></div> <!> <div class="rounded-xl border border-gray-200 bg-white shadow-sm dark:border-gray-800 dark:bg-gray-900 overflow-hidden"><div class="border-b border-gray-100 px-6 py-4 flex items-center justify-between dark:border-gray-800"><div><h2 class="text-sm font-bold text-gray-900 dark:text-white">HTTP API Quickstart</h2> <p class="text-xs text-gray-500 dark:text-gray-400">Endpoint: <code class="text-teal-600 dark:text-teal-400">POST https://sms.whiteinfobd.com/api/v1/messages</code></p></div> <div class="flex items-center gap-1 rounded-lg bg-gray-100 p-1 dark:bg-gray-800"></div></div> <div class="p-6 bg-gray-950 font-mono text-xs text-gray-300 overflow-x-auto"><!></div></div></div> <!>',1);function Re(L,z){le(z,!0);let g=p(ce([])),S=p(!0),b=p(!1),y=p(""),w=p(!1),x=p(null),u=p("curl");async function $(){r(S,!0);try{const e=await K.get("/api/v1/customer/tokens");r(g,e.data,!0)}catch{r(g,[{id:1,name:"Production Server REST API",last_used_at:new Date(Date.now()-12e4).toISOString(),expires_at:null,created_at:new Date(Date.now()-30*864e5).toISOString()}],!0)}finally{r(S,!1)}}de(()=>$());async function J(){if(!a(y).trim()){k.error("Validation error","Token name is required.");return}r(w,!0);try{const e=await K.post("/api/v1/customer/tokens",{name:a(y)});r(x,e.token,!0),r(g,[e.data,...a(g)],!0),k.success("Token Generated","Make sure to copy your new token now."),r(y,"")}catch(e){const t=e instanceof Error?e.message:"Failed to generate token";k.error("Token Generation Failed",t)}finally{r(w,!1)}}async function Q(e){try{await navigator.clipboard.writeText(e),k.success("Copied!","Token copied to clipboard.")}catch{k.info("Copy",e)}}const V=[{key:"name",label:"Token Name",class:"font-semibold text-gray-100"},{key:"last_used_at",label:"Last Used",render:e=>e.last_used_at?new Date(e.last_used_at).toLocaleString():"Never Used",class:"text-xs text-gray-400 font-mono"},{key:"expires_at",label:"Expires",render:e=>e.expires_at?new Date(e.expires_at).toLocaleDateString():"Never Expires",class:"text-xs text-gray-400"},{key:"created_at",label:"Created",render:e=>new Date(e.created_at).toLocaleDateString()}];var H=Pe();ye("12rvbtq",e=>{ge(()=>{ue.title="Developer API Access – SMS Whiteinfo"})});var C=R(H),I=n(C),X=i(n(I),2);o(I);var Y=i(I,2);ke(Y,{get columns(){return V},get rows(){return a(g)},get loading(){return a(S)},get total(){return a(g).length},perPage:15,currentPage:1});var B=i(Y,2),A=n(B),F=i(n(A),2);be(F,20,()=>["curl","php","python","node"],e=>e,(e,t)=>{var s=we(),D=j(s,!0);M(()=>{fe(s,1,`rounded-md px-2.5 py-1 text-xs font-semibold uppercase transition ${a(u)===t?"bg-teal-600 text-white shadow-sm":"text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200"}`),U(D,t)}),h("click",s,()=>r(u,t,!0)),d(e,s)}),o(F),o(A);var G=i(A,2),Z=n(G);{var ee=e=>{var t=P(),s=n(t);s.textContent=`curl -X POST "https://sms.whiteinfobd.com/api/v1/messages" \\
  -H "Authorization: Bearer YOUR_API_TOKEN" \\
  -H "Content-Type: application/json" \\
  -H "Accept: application/json" \\
  -d '{
    "sender_id": "WHITEINFO",
    "to": "+8801700000000",
    "message": "Your OTP code is 951234."
  }'`,o(t),d(e,t)},te=e=>{var t=P(),s=n(t);s.textContent=`$response = Http::withToken('YOUR_API_TOKEN')
    ->post('https://sms.whiteinfobd.com/api/v1/messages', [
        'sender_id' => 'WHITEINFO',
        'to' => '+8801700000000',
        'message' => 'Your OTP code is 951234.',
    ]);

return $response->json();`,o(t),d(e,t)},ae=e=>{var t=P(),s=n(t);s.textContent=`import requests

url = "https://sms.whiteinfobd.com/api/v1/messages"
headers = {
    "Authorization": "Bearer YOUR_API_TOKEN",
    "Content-Type": "application/json"
}
payload = {
    "sender_id": "WHITEINFO",
    "to": "+8801700000000",
    "message": "Your OTP code is 951234."
}

res = requests.post(url, json=payload, headers=headers)
print(res.json())`,o(t),d(e,t)},re=e=>{var t=P(),s=n(t);s.textContent=`const res = await fetch('https://sms.whiteinfobd.com/api/v1/messages', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_TOKEN',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    sender_id: 'WHITEINFO',
    to: '+8801700000000',
    message: 'Your OTP code is 951234.'
  })
});

const data = await res.json();
console.log(data);`,o(t),d(e,t)};W(Z,e=>{a(u)==="curl"?e(ee):a(u)==="php"?e(te,1):a(u)==="python"?e(ae,2):a(u)==="node"&&e(re,3)})}o(G),o(B),o(C);var oe=i(C,2);_e(oe,{title:"Generate API Token",description:"Create a personal Bearer access key for your application.",get open(){return a(b)},set open(e){r(b,e,!0)},children:(e,t)=>{var s=ie(),D=R(s);{var se=v=>{var l=Te(),c=n(l),f=i(n(c),4),m=n(f);q(m);var T=i(m,2);o(f),o(c);var O=i(c,2),N=j(O);o(l),M(()=>xe(m,a(x))),h("click",T,()=>Q(a(x))),h("click",N,()=>r(b,!1)),d(v,l)},ne=v=>{var l=Oe(),c=n(l),f=i(n(c),2);q(f),o(c);var m=i(c,2),T=n(m),O=i(T,2),N=j(O,!0);o(m),o(l),M(()=>{O.disabled=a(w),U(N,a(w)?"Generating...":"Create Token")}),me("submit",l,E=>{E.preventDefault(),J()}),he(f,()=>a(y),E=>r(y,E)),h("click",T,()=>r(b,!1)),d(v,l)};W(D,v=>{a(x)?v(se):v(ne,-1)})}d(e,s)},$$slots:{default:!0}}),h("click",X,()=>{r(x,null),r(b,!0)}),d(L,H),pe()}ve(["click"]);export{Re as component};
