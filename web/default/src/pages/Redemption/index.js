import React from 'react';
import { Segment, Header } from 'semantic-ui-react';
import RedemptionsTable from '../../components/RedemptionsTable';

const Redemption = () => (
  <>
    <Segment>
      <Header as='h3'>Manage Redeem Codes</Header>
      <RedemptionsTable/>
    </Segment>
  </>
);

export default Redemption;
