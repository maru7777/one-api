import React from 'react';
import { Message } from 'semantic-ui-react';

const NotFound = () => (
  <>
    <Message negative>
      <Message.Header>Page does not exist</Message.Header>
      <p>Please check if your browser address is correct</p>
    </Message>
  </>
);

export default NotFound;
