package internal

const tmpl = `<!doctype html>
<head>
<title>My coverage</title>
<style>
    div { border: 1px solid transparent; }
    .table { border: 1px solid #aaa; border-collapse: collapse; }
    .source { display: none; }
    .visible { display: block; }
    td { border: 1px solid #aaa; padding: 5px; }
    a { text-decoration: none; }
    .ok { background: rgb(233,245,212); }
    .ok2 { background: rgb(94,145,53); }
    .warn { background: rgb(253,245,200); }
    .warn2 { background: rgb(242,202,83); }
    .error { background: pink; }
    .progress { width: 100px; height: 20px; }
    .progress > div { height: calc(20px - 1px); }
    .ok .progress { border: 1px solid rgb(94,145,53); }
    .ok .progress > div { background: rgb(94,145,53); }
    .warn .progress { border: 1px solid rgb(242,202,83); }
    .warn .progress > div { background: rgb(242,202,83); }
    .error .progress { border: 1px solid darkred; }
    .error .progress > div { background: darkred; }
</style>
</head>
<body>
<div class="breadcrumbs"></div>
<div class="stats"></div>
<table class="table"></table>
<!-- REPORT -->
<!-- SOURCE -->
<script>
<!-- SCRIPT -->
</script>
</body>
`
