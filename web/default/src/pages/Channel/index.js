import React from 'react';
import { Header, Segment } from 'semantic-ui-react';
import ChannelsTable from '../../components/ChannelsTable';

const Channel = () => (
  <>
    <Segment>
      <Header as='h3'>Manage Channels</Header>
      <ChannelsTable />
    </Segment>
  </>
);

export default Channel;
