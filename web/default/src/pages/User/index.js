import React from 'react';
import { Card } from 'semantic-ui-react';
import UsersTable from '../../components/UsersTable';

const User = () => (
  <div className='dashboard-container'>
    <Card fluid className='chart-card'>
      <Card.Content>
        <Card.Header className='header'>Manage Users</Card.Header>
        <UsersTable />
      </Card.Content>
    </Card>
  </div>
);

export default User;
