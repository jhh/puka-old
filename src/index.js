
import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Router, Route, IndexRoute, browserHistory } from 'react-router';
import { Provider } from 'react-redux';
import { createStore, applyMiddleware, compose } from 'redux';
import thunkMiddleware from 'redux-thunk';

import rootReducer from './reducers';
import MainContainer from './components/main';
import BookmarkPage from './components/bookmark-page';
import BookmarkForm from './components/bookmark-form';

require('file?name=[name].[ext]!../assets/favicon.ico');
require('../assets/octicons/octicons.css');
require('../assets/octicons/octicons.eot');
require('../assets/octicons/octicons.woff');
require('../assets/octicons/octicons.ttf');
require('../assets/octicons/octicons.svg');

const store = createStore(
  rootReducer,
  compose(
    applyMiddleware(thunkMiddleware),
    process.env.NODE_ENV !== 'production' && window.devToolsExtension
      ? window.devToolsExtension() : f => f
  )
);

const routes = (
  <Provider store={store}>
    <Router history={browserHistory}>
      <Route path="/" component={MainContainer}>
        <IndexRoute component={BookmarkPage} />
        <Route path="tag/:tag" component={BookmarkPage} />
        <Route path="new" component={BookmarkForm} />
      </Route>
    </Router>
  </Provider>
);

ReactDOM.render(routes, document.getElementById('app'));
