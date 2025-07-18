import React, { useState } from 'react';
import { Button, Modal, Header, Segment, Label, Message, List } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';

const ChannelDebugPanel = ({ channelId, channelType, channelName }) => {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [migrationStatus, setMigrationStatus] = useState(null);

  const checkMigrationStatus = async () => {
    setLoading(true);
    try {
      const res = await API.get(`/api/debug/channel/${channelId}/migration-status`);
      if (res.data.success) {
        setMigrationStatus(res.data.data);
      } else {
        showError('Failed to get migration status: ' + res.data.message);
      }
    } catch (error) {
      showError('Failed to check migration status: ' + error.message);
    }
    setLoading(false);
  };

  const fixChannel = async () => {
    setLoading(true);
    try {
      const res = await API.post(`/api/debug/channel/${channelId}/fix`);
      if (res.data.success) {
        showSuccess('Channel fixed successfully. Please refresh the page.');
        await checkMigrationStatus(); // Refresh status
      } else {
        showError('Failed to fix channel: ' + res.data.message);
      }
    } catch (error) {
      showError('Failed to fix channel: ' + error.message);
    }
    setLoading(false);
  };

  const debugChannel = async () => {
    setLoading(true);
    try {
      const res = await API.post(`/api/debug/channel/${channelId}/debug`);
      if (res.data.success) {
        showSuccess('Debug information logged. Check application logs.');
      } else {
        showError('Failed to debug channel: ' + res.data.message);
      }
    } catch (error) {
      showError('Failed to debug channel: ' + error.message);
    }
    setLoading(false);
  };

  const getMigrationStatusColor = (status) => {
    switch (status) {
      case 'migrated': return 'green';
      case 'migrated_with_legacy': return 'yellow';
      case 'needs_migration': return 'red';
      case 'empty': return 'grey';
      default: return 'grey';
    }
  };

  const getMigrationStatusText = (status) => {
    switch (status) {
      case 'migrated': return 'Fully Migrated';
      case 'migrated_with_legacy': return 'Migrated (Legacy Data Present)';
      case 'needs_migration': return 'Needs Migration';
      case 'empty': return 'No Pricing Data';
      default: return 'Unknown';
    }
  };

  return (
    <>
      <Button 
        size="mini" 
        color="blue" 
        onClick={() => {
          setOpen(true);
          checkMigrationStatus();
        }}
        style={{ marginLeft: '10px' }}
      >
        Debug
      </Button>

      <Modal open={open} onClose={() => setOpen(false)} size="small">
        <Header icon="bug" content={`Channel Debug: ${channelName} (ID: ${channelId})`} />
        <Modal.Content>
          <Segment loading={loading}>
            <Header as="h4">Migration Status</Header>
            
            {migrationStatus && (
              <>
                <Label color={getMigrationStatusColor(migrationStatus.migration_status)} size="large">
                  {getMigrationStatusText(migrationStatus.migration_status)}
                </Label>
                
                <div style={{ marginTop: '15px' }}>
                  <strong>Channel Info:</strong>
                  <List>
                    <List.Item>ID: {migrationStatus.channel_id}</List.Item>
                    <List.Item>Name: {migrationStatus.channel_name}</List.Item>
                    <List.Item>Type: {migrationStatus.channel_type}</List.Item>
                  </List>
                </div>

                <div style={{ marginTop: '15px' }}>
                  <strong>Data Status:</strong>
                  <List>
                    <List.Item>
                      <Label color={migrationStatus.has_model_configs ? 'green' : 'red'} size="mini">
                        {migrationStatus.has_model_configs ? 'YES' : 'NO'}
                      </Label>
                      Has Model Configs (Unified Format)
                      {migrationStatus.model_configs_count && (
                        <span> - {migrationStatus.model_configs_count} models</span>
                      )}
                    </List.Item>
                    <List.Item>
                      <Label color={migrationStatus.has_model_ratio ? 'yellow' : 'grey'} size="mini">
                        {migrationStatus.has_model_ratio ? 'YES' : 'NO'}
                      </Label>
                      Has Model Ratio (Legacy)
                      {migrationStatus.model_ratio_count && (
                        <span> - {migrationStatus.model_ratio_count} models</span>
                      )}
                    </List.Item>
                    <List.Item>
                      <Label color={migrationStatus.has_completion_ratio ? 'yellow' : 'grey'} size="mini">
                        {migrationStatus.has_completion_ratio ? 'YES' : 'NO'}
                      </Label>
                      Has Completion Ratio (Legacy)
                    </List.Item>
                  </List>
                </div>

                {migrationStatus.model_configs_models && (
                  <div style={{ marginTop: '15px' }}>
                    <strong>Models in Unified Config:</strong>
                    <div style={{ marginTop: '5px' }}>
                      {migrationStatus.model_configs_models.map(model => (
                        <Label key={model} size="mini" style={{ margin: '2px' }}>
                          {model}
                        </Label>
                      ))}
                    </div>
                  </div>
                )}

                {migrationStatus.model_ratio_models && (
                  <div style={{ marginTop: '15px' }}>
                    <strong>Models in Legacy Ratio:</strong>
                    <div style={{ marginTop: '5px' }}>
                      {migrationStatus.model_ratio_models.map(model => (
                        <Label key={model} size="mini" color="yellow" style={{ margin: '2px' }}>
                          {model}
                        </Label>
                      ))}
                    </div>
                  </div>
                )}

                {migrationStatus.migration_status === 'needs_migration' && (
                  <Message warning>
                    <Message.Header>Migration Needed</Message.Header>
                    <p>This channel has legacy pricing data that needs to be migrated to the unified format.</p>
                  </Message>
                )}

                {migrationStatus.migration_status === 'migrated_with_legacy' && (
                  <Message info>
                    <Message.Header>Migration Complete with Legacy Data</Message.Header>
                    <p>This channel has been migrated but still contains legacy data for backward compatibility.</p>
                  </Message>
                )}
              </>
            )}
          </Segment>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={debugChannel} loading={loading}>
            Log Debug Info
          </Button>
          <Button onClick={checkMigrationStatus} loading={loading}>
            Refresh Status
          </Button>
          {migrationStatus && migrationStatus.migration_status === 'needs_migration' && (
            <Button color="orange" onClick={fixChannel} loading={loading}>
              Fix Channel
            </Button>
          )}
          <Button onClick={() => setOpen(false)}>
            Close
          </Button>
        </Modal.Actions>
      </Modal>
    </>
  );
};

export default ChannelDebugPanel;
