"use strict";

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

function renderFileList(files, dispatch) {
    return function(json) {
        if (json.Docs == null || json.Docs.length < 1) {
            m.render(files, []);
            return;
        }
        let a = []
        let i = 0;
        for(let filename of json.Docs) {
            if (i++ > 18) break;
            a.push(renderFileItem(filename, dispatch));
        }
        let panel = m('nav', {class: 'panel has-background-white'}, a);
        m.render(files, panel);
    }
}

const maxWidth = 60;
const leftSize = (maxWidth-3)/3;
const rightSize = leftSize * 2;

function renderFileItem(filename, dispatch) {
    let label = filename;
    if (filename.length > maxWidth) {
        let start = filename.slice(0, leftSize);
        let pos =  filename.length - rightSize;
        let end = filename.slice(pos, filename.length);
        label = start + "..." + end;
    }
    return m("a", {class: "panel-block", onclick: e => dispatch(setFile(filename))}, [m("span", {class: "panel-icon"}, m("i", {class: "fab fa-java"}, "")), label])
}

function renderBreadcrumbs(breadcrumbs) {
    return function(filename) {
        if (filename == null || filename === '') {
            m.render(breadcrumbs, []);
            return;
        }
        let items = [];
        let crumbs = filename.split('/');
        for (let i = 0; i < crumbs.length -1; i++) {
            let li = m('li', m('a', {}, crumbs[i]));
            items.push(li)
        }
        let last = crumbs.length - 1;
        let li = m('li', {class: 'is-active'}, m('a', {'aria-current': 'page'}, crumbs[last]));
        items.push(li)
        m.render(breadcrumbs, items);
    }
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

function renderQueryState(el) {
    return isQuerying => {
        if (isQuerying) {
            m.render(el, m('i', { class: 'fas fa-spinner', 'aria-hidden': true }, ""));
            return;
        }
        m.render(el, m('i', { class: 'fas fa-search', 'aria-hidden': true }, ""));
    };
}

function renderCode(code, highlight) {
    return (file) => {
        let name = file.name;
        let content = file.content;
        if (name == null || name === '') {
            return;
        }
        let segments = name.split('.');
        let last = segments.length - 1;
        let className = 'line-numbers  language-' + segments[last];
        code.parentElement.className =  className;
        code.innerHTML = content || '';
        if (content == null) {
            return;
        }
        highlight();
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

    search.addEventListener('input', dispatchQueryTerm);
    search.addEventListener('focus', dispatchQueryTerm);

    regSub(store, ['file'], renderCode(code, Prism.highlightAll));
    regSub(store, ['file', 'name'], renderBreadcrumbs(breadcrumbs));
    regSub(store, ['file', 'name'], dispatchClear);
    regSub(store, ['file', 'name'], fetchFile);
    regSub(store, ['query', 'isQuerying'], renderQueryState(searchSpinner));
    regSub(store, ['query', 'result'], renderFileList(files, store.dispatch));
    regSub(store, ['query', 'term'], query);

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
