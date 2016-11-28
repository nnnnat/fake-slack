import React from 'react';
import { render } from 'react-dom';
import App from './components/App.jsx';

const mainEl = document.querySelector('main');

render(<App />, mainEl);