import React from 'react';
import { Card } from 'semantic-ui-react';
import RedemptionsTable from '../../components/RedemptionsTable';

const Redemption = () => (
  <div className='dashboard-container'>
    <Card fluid className='chart-card'>
      <Card.Content>
        <Card.Header className='header'>Manage Redeem Code</Card.Header>
        <RedemptionsTable />
      </Card.Content>
    </Card>
  </div>
);

export default Redemption;
