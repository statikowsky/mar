---
title: 'Drop custom directives: GFM alerts + mar migrate directives'
status: active
created: "2026-06-12T19:26:05.529064Z"
updated: "2026-06-12T19:37:36.626528Z"
---
Implement DOC-DIRECTIVES: remove the ::: directive preprocessor, render GFM alerts via a goldmark AST transformer, add fence-aware 'mar migrate directives' (--dry-run) backed by store.RewriteBodies, update the HTML importer, CSS, README and guide, then migrate this repo's own docs.
