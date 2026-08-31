 function syntaxHighlightJson(json) {
  if (typeof json !== 'string') {
    json = JSON.stringify(json, null, 2);
  }
  json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
    let cls = 'text-amber-400'; 
    if (/^"/.test(match)) {
      if (/:$/.test(match)) {
        cls = 'text-sky-400 font-semibold';
      } else {
        cls = 'text-emerald-300';  
      }
    } else if (/true|false/.test(match)) {
      cls = 'text-purple-400 font-bold';  
    } else if (/null/.test(match)) {
      cls = 'text-rose-400 font-bold'; 
    }
    return '<span class="' + cls + '">' + match + '</span>';
  });
}
