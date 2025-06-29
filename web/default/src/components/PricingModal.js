import React, { useState, useEffect } from 'react';
import { Modal, Form, Button, Message, Table, Input, Icon, Dropdown } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';
import { useTranslation } from 'react-i18next';

const PricingModal = ({ open, onClose, channelId, channelName, channelType }) => {
  const { t } = useTranslation();
  const [modelRatio, setModelRatio] = useState({});
  const [completionRatio, setCompletionRatio] = useState({});
  const [loading, setLoading] = useState(false);
  const [newModelName, setNewModelName] = useState('');
  const [newModelPrice, setNewModelPrice] = useState('');
  const [newCompletionName, setNewCompletionName] = useState('');
  const [newCompletionPrice, setNewCompletionPrice] = useState('');
  const [supportedModels, setSupportedModels] = useState([]);

  useEffect(() => {
    if (open && channelId) {
      loadPricing();
      loadSupportedModels();
    }
  }, [open, channelId]);

  const loadSupportedModels = async () => {
    try {
      const response = await API.get('/api/models');
      if (response.data.success) {
        const channelModels = response.data.data[channelType] || [];
        setSupportedModels(channelModels.map(model => ({ key: model, value: model, text: model })));
      }
    } catch (error) {
      console.error('Failed to load supported models:', error);
    }
  };

  const loadPricing = async () => {
    setLoading(true);
    try {
      const response = await API.get(`/api/channel/pricing/${channelId}`);
      if (response.data.success) {
        setModelRatio(response.data.data.model_ratio || {});
        setCompletionRatio(response.data.data.completion_ratio || {});
      } else {
        showError(response.data.message);
      }
    } catch (error) {
      showError('Failed to load pricing data');
    } finally {
      setLoading(false);
    }
  };

  const savePricing = async () => {
    setLoading(true);
    try {
      const response = await API.put(`/api/channel/pricing/${channelId}`, {
        model_ratio: Object.keys(modelRatio).length > 0 ? modelRatio : null,
        completion_ratio: Object.keys(completionRatio).length > 0 ? completionRatio : null
      });
      if (response.data.success) {
        showSuccess('Pricing updated successfully');
        onClose();
      } else {
        showError(response.data.message);
      }
    } catch (error) {
      showError('Failed to save pricing data');
    } finally {
      setLoading(false);
    }
  };

  const addModelRatio = () => {
    if (newModelName && newModelPrice) {
      setModelRatio(prev => ({
        ...prev,
        [newModelName]: parseFloat(newModelPrice)
      }));
      setNewModelName('');
      setNewModelPrice('');
    }
  };

  const addCompletionRatio = () => {
    if (newCompletionName && newCompletionPrice) {
      setCompletionRatio(prev => ({
        ...prev,
        [newCompletionName]: parseFloat(newCompletionPrice)
      }));
      setNewCompletionName('');
      setNewCompletionPrice('');
    }
  };

  const removeModelRatio = (modelName) => {
    setModelRatio(prev => {
      const newRatio = { ...prev };
      delete newRatio[modelName];
      return newRatio;
    });
  };

  const removeCompletionRatio = (modelName) => {
    setCompletionRatio(prev => {
      const newRatio = { ...prev };
      delete newRatio[modelName];
      return newRatio;
    });
  };

  const updateModelRatio = (modelName, value) => {
    setModelRatio(prev => ({
      ...prev,
      [modelName]: parseFloat(value) || 0
    }));
  };

  const updateCompletionRatio = (modelName, value) => {
    setCompletionRatio(prev => ({
      ...prev,
      [modelName]: parseFloat(value) || 0
    }));
  };

  return (
    <Modal open={open} onClose={onClose} size="large">
      <Modal.Header>
        <Icon name="dollar sign" />
        Channel Pricing - {channelName}
      </Modal.Header>
      <Modal.Content scrolling>
        <Message info>
          <Message.Header>Channel-Specific Pricing</Message.Header>
          <p>
            Configure channel-specific pricing here. If a model is not listed, the global pricing will be used as fallback.
            Leave empty to use global pricing for all models.
          </p>
        </Message>

        {/* Model Ratio Section */}
        <Form.Field>
          <label><Icon name="settings" /> Model Pricing Ratios</label>
          <Message size="small">
            Set custom pricing for specific models on this channel
          </Message>
        </Form.Field>

        <Form>
          <Form.Group widths="equal">
            <Form.Field>
              <label>Model Name</label>
              <Dropdown
                placeholder="Select or type model name"
                fluid
                search
                selection
                allowAdditions
                value={newModelName}
                options={supportedModels}
                onAddItem={(e, { value }) => {
                  setSupportedModels(prev => [...prev, { key: value, value, text: value }]);
                }}
                onChange={(e, { value }) => setNewModelName(value)}
              />
            </Form.Field>
            <Form.Field>
              <label>Price Ratio</label>
              <Input
                type="number"
                step="0.000001"
                placeholder="e.g., 0.03"
                value={newModelPrice}
                onChange={(e) => setNewModelPrice(e.target.value)}
              />
            </Form.Field>
            <Form.Field>
              <label>&nbsp;</label>
              <Button
                primary
                onClick={addModelRatio}
                disabled={!newModelName || !newModelPrice}
              >
                <Icon name="plus" /> Add
              </Button>
            </Form.Field>
          </Form.Group>
        </Form>

        {Object.keys(modelRatio).length > 0 && (
          <Table celled>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>Model Name</Table.HeaderCell>
                <Table.HeaderCell>Price Ratio</Table.HeaderCell>
                <Table.HeaderCell>Actions</Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {Object.entries(modelRatio).map(([modelName, ratio]) => (
                <Table.Row key={modelName}>
                  <Table.Cell>{modelName}</Table.Cell>
                  <Table.Cell>
                    <Input
                      type="number"
                      step="0.000001"
                      value={ratio}
                      onChange={(e) => updateModelRatio(modelName, e.target.value)}
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Button
                      size="mini"
                      negative
                      onClick={() => removeModelRatio(modelName)}
                    >
                      <Icon name="trash" /> Remove
                    </Button>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        )}

        {/* Completion Ratio Section */}
        <Form.Field style={{ marginTop: '2em' }}>
          <label><Icon name="settings" /> Completion Pricing Ratios</label>
          <Message size="small">
            Set custom completion token pricing multipliers for specific models
          </Message>
        </Form.Field>

        <Form>
          <Form.Group widths="equal">
            <Form.Field>
              <label>Model Name</label>
              <Dropdown
                placeholder="Select or type model name"
                fluid
                search
                selection
                allowAdditions
                value={newCompletionName}
                options={supportedModels}
                onAddItem={(e, { value }) => {
                  setSupportedModels(prev => [...prev, { key: value, value, text: value }]);
                }}
                onChange={(e, { value }) => setNewCompletionName(value)}
              />
            </Form.Field>
            <Form.Field>
              <label>Completion Ratio</label>
              <Input
                type="number"
                step="0.1"
                placeholder="e.g., 2.0"
                value={newCompletionPrice}
                onChange={(e) => setNewCompletionPrice(e.target.value)}
              />
            </Form.Field>
            <Form.Field>
              <label>&nbsp;</label>
              <Button
                primary
                onClick={addCompletionRatio}
                disabled={!newCompletionName || !newCompletionPrice}
              >
                <Icon name="plus" /> Add
              </Button>
            </Form.Field>
          </Form.Group>
        </Form>

        {Object.keys(completionRatio).length > 0 && (
          <Table celled>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>Model Name</Table.HeaderCell>
                <Table.HeaderCell>Completion Ratio</Table.HeaderCell>
                <Table.HeaderCell>Actions</Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {Object.entries(completionRatio).map(([modelName, ratio]) => (
                <Table.Row key={modelName}>
                  <Table.Cell>{modelName}</Table.Cell>
                  <Table.Cell>
                    <Input
                      type="number"
                      step="0.1"
                      value={ratio}
                      onChange={(e) => updateCompletionRatio(modelName, e.target.value)}
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Button
                      size="mini"
                      negative
                      onClick={() => removeCompletionRatio(modelName)}
                    >
                      <Icon name="trash" /> Remove
                    </Button>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        )}
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={onClose} disabled={loading}>
          <Icon name="cancel" /> Cancel
        </Button>
        <Button primary onClick={savePricing} loading={loading}>
          <Icon name="save" /> Save Pricing
        </Button>
      </Modal.Actions>
    </Modal>
  );
};

export default PricingModal;
