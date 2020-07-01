"use strict";

Prism.plugins.autoloader.loadLanguages(['java', 'go', 'javascript']);

const CLEAR_QUERY = 'CLEAR_QUERY';
const FETCH_QUERY_RESULT = 'FETCH_QUERY_RESULT ';
const SET_QUERY_TERM = 'SET_QUERY';
const SET_QUERY_RESULT = 'SET_QUERY_RESULT';
const SET_FILE = 'SET_FILE';
const SET_FILE_CONTENT = 'SET_FILE_CONTENT';
const INITIAL_QUERY = { isQuerying: false, result: {} };

function queryReducer(state = INITIAL_QUERY, action) {
    switch (action.type) {
        case CLEAR_QUERY:
            return Object.assign({}, state, { term: null, result: {}, isQuerying: false });

        case FETCH_QUERY_RESULT:
            return Object.assign({}, state, { isQuerying: true });

        case SET_QUERY_TERM:
            return Object.assign({}, state, { term: action.value });

        case SET_QUERY_RESULT:
            return Object.assign({}, state, { result: action.value, isQuerying: false });

        default:
            return state;
    }
}

function fileReducer(state = {}, action) {
    switch (action.type) {
        case SET_FILE:
            return Object.assign({}, state, { name: action.value, content: null });

        case SET_FILE_CONTENT:
            return Object.assign({}, state, { content: action.value });

        default:
            return state
    }
}

let Breadcrumbs = {
    view: function(vnode) {
        let {segments} = vnode.attrs;
        return segments.map((c, i) => {
            if (i === segments.length - 1) {
                return m("li", {"class":"breadcrumb-item","aria-current":"page"}, c);
            }
            return m("li", {"class":"breadcrumb-item"}, m("a", {"href":"#"}, c));
        });
    }
}

let CodeBlock = {
    view: function(vnode) {
        let {className, html} = vnode.attrs;
        return m("pre", {"class": "line-numbers"},
            m("code", {"class": className}, [m.trust(html)])
        );
    },
}

let FileList = {
    view: function (vnode) {
        let {docs, dispatch} = vnode.attrs;
        return docs.map((d) => {
            let filename = d.Document;
            let key = filename;
            let label = toLabel(filename);
            return m(FileItem, {dispatch, filename, key, label})
        });
    }
}

let FileItem = {
    view: function (vnode) {
        let {dispatch, filename, label} = vnode.attrs;
        return m("li", {class: "autocomplete-item", onclick: e => dispatch(setFile(filename))},
            label);
    }
}

let ProgressIndicator = {
    view: function (vnode) {
        let c = vnode.attrs.isQuerying ? 'fas fa-dumpster-fire' : 'fas fa-dumpster';
        return m('i', { class: c, 'aria-hidden': true }, "");
    }
}

const maxWidth = 90;
const leftSize = (maxWidth-3)/3;
const rightSize = leftSize * 2;

function toLabel(filename) {
    if (filename.length > maxWidth) {
        let start = filename.slice(0, leftSize);
        let pos =  filename.length - rightSize;
        let end = filename.slice(pos, filename.length);
        return start + "..." + end;
    }
    return filename;
}

function setFile(value) {
    return {
        type: SET_FILE,
        value
    };
}

function setQueryTerm(value) {
    return {
        type: SET_QUERY_TERM,
        value
    };
}

function setQueryResult(value) {
    return {
        type: SET_QUERY_RESULT,
        value
    };
}

function clearQuery() {
    return {
        type: CLEAR_QUERY,
    }
}

function fetchQueryResult() {
    return {
        type: FETCH_QUERY_RESULT,
    }
}

function fileContent(value) {
    return {
        type: SET_FILE_CONTENT,
        value
    };
}

function regSub(store, path, fn) {
    let prev = null;
    store.subscribe(() => {
        let v = store.getState();
        for (let k of path) {
            if (v == null) break;
            v = v[k];
        }
        if (v === prev) {
            return;
        }
        prev = v;
        fn(v);
    });
}

function renderFileList(el, dispatch) {
    return function(json) {
        let docs = json.Docs;
        if (json.Docs == null) {
            docs = [];
        }
        if (docs.length > 18) {
            docs = docs.slice(0, 18);
        }
        m.render(el, m(FileList, {dispatch, docs}));
    }
}

function renderBreadcrumbs(el) {
    return function(filename) {
        if (filename == null || filename === '') {
            m.render(el, m(Breadcrumbs, {segments: []}));
            return;
        }
        m.render(el, m(Breadcrumbs, {segments: filename.split('/')}));
    }
}

function renderQueryState(el) {
    return isQuerying => {
        m.render(el, m(ProgressIndicator, {isQuerying}));
    };
}

function renderCode(el) {
    return ({content, name}) => {
        if (name == null || name === '' || content == null) {
            return;
        }
        let segments = name.split('.');
        let lang = segments[segments.length - 1];
        let className = 'language-' + lang;
        let g = Prism.languages[lang];
        if (g == null) {
            return;
        }
        let html = Prism.highlight(content, g, "");
        m.render(el, m(CodeBlock, {className, html}));
    }
}

function Exec(breadcrumbs, code, files, search, searchSpinner) {

    let rootReducer = Redux.combineReducers({
        query: queryReducer,
        file: fileReducer,
    });
    let store = Redux.createStore(rootReducer);

    let dispatchClear = () => store.dispatch(clearQuery());
    let dispatchQueryTerm = e => store.dispatch(setQueryTerm(e.target.value));
    let query = (v) => {
        if (v == null) return;
        store.dispatch(fetchQueryResult());
        fetch('/search?q='+encodeURIComponent(v))
            .then(response => response.json())
            .then(json => store.dispatch(setQueryResult(json))) };
    let fetchFile = (v) => {
        if (v == null) return;
        fetch('/files/'+v)
            .then(response => response.text())
            .then(text => store.dispatch(fileContent(text))) };

    let setLocationHash = (v) => {
      if (v == null) return;
      location.hash = encodeURIComponent(v);
    };

    let lastHash = null;
    let getLocationHash = () => {
        if (location.hash === lastHash) {
            return;
        }
        lastHash = location.hash;
        if (location.hash.length < 2) {
            return;
        }
        let filename = decodeURIComponent(location.hash.slice(1));
        store.dispatch(setFile(filename));
    };

    search.addEventListener('input', dispatchQueryTerm);
    search.addEventListener('focus', dispatchQueryTerm);
    window.addEventListener('hashchange', getLocationHash);

    regSub(store, ['file'], renderCode(code));
    regSub(store, ['file', 'name'], renderBreadcrumbs(breadcrumbs));
    regSub(store, ['file', 'name'], dispatchClear);
    regSub(store, ['file', 'name'], fetchFile);
    regSub(store, ['file', 'name'], setLocationHash)
    regSub(store, ['query', 'isQuerying'], renderQueryState(searchSpinner));
    regSub(store, ['query', 'result'], renderFileList(files, store.dispatch));
    regSub(store, ['query', 'term'], query);

    getLocationHash()
    search.focus();
}

function main() {
    const breadcrumbs = document.getElementById('breadcrumbs');
    const code = document.getElementById('code');
    const files = document.getElementById('files');
    const search = document.getElementById('search');
    const searchSpinner = document.getElementById('searchSpinner')
    Exec(breadcrumbs, code, files, search, searchSpinner)
}

window.onload = main;
