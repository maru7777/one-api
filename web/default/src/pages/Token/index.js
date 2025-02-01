import React from 'react';
import { Card } from 'semantic-ui-react';
import TokensTable from '../../components/TokensTable';

const Token = () => (
  <div className='dashboard-container'>
    <Card fluid className='chart-card'>
      <Card.Content>
        <Card.Header className='header'>Apikeys</Card.Header>
        <TokensTable />
      </Card.Content>
    </Card>
  </div>
);

export default Token;
