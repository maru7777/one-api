import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Input, Table, Header, Segment, Message, Loader, Dropdown, Grid, Button } from 'semantic-ui-react';
import { API, showError } from '../../helpers';

const Models = () => {
  const { t } = useTranslation();
  const [modelsData, setModelsData] = useState({});
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedChannels, setSelectedChannels] = useState([]);
  const [filteredData, setFilteredData] = useState({});

  const fetchModelsData = async () => {
    try {
      setLoading(true);
      const res = await API.get('/api/models/display');
      const { success, message, data } = res.data;
      if (success) {
        setModelsData(data || {});
        setFilteredData(data || {});
      } else {
        showError(message);
      }
    } catch (error) {
      if (error.response && error.response.status === 401) {
        showError('Please login to view available models');
      } else {
        showError('Failed to fetch models data');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchModelsData();
  }, []);

  useEffect(() => {
    let filtered = { ...modelsData };

    // Filter by selected channels
    if (selectedChannels.length > 0) {
      const channelFiltered = {};
      selectedChannels.forEach(channelName => {
        if (filtered[channelName]) {
          channelFiltered[channelName] = filtered[channelName];
        }
      });
      filtered = channelFiltered;
    }

    // Filter by search term
    if (searchTerm) {
      const searchFiltered = {};
      Object.keys(filtered).forEach(channelName => {
        const channelData = filtered[channelName];
        const filteredModels = {};

        Object.keys(channelData.models).forEach(modelName => {
          if (modelName.toLowerCase().includes(searchTerm.toLowerCase())) {
            filteredModels[modelName] = channelData.models[modelName];
          }
        });

        if (Object.keys(filteredModels).length > 0) {
          searchFiltered[channelName] = {
            ...channelData,
            models: filteredModels
          };
        }
      });
      filtered = searchFiltered;
    }

    setFilteredData(filtered);
  }, [searchTerm, selectedChannels, modelsData]);

  const formatPrice = (price) => {
    if (price === 0) return 'Free';
    if (price < 0.001) return `$${price.toFixed(6)}`;
    if (price < 1) return `$${price.toFixed(4)}`;
    return `$${price.toFixed(2)}`;
  };

  const formatMaxTokens = (maxTokens) => {
    if (maxTokens === 0) return 'Unlimited';
    if (maxTokens >= 1000000) return `${(maxTokens / 1000000).toFixed(1)}M`;
    if (maxTokens >= 1000) return `${(maxTokens / 1000).toFixed(0)}K`;
    return maxTokens.toString();
  };

  const renderChannelModels = (channelName, channelInfo) => {
    const models = Object.keys(channelInfo.models).map(modelName => ({
      model: modelName,
      inputPrice: channelInfo.models[modelName].input_price,
      outputPrice: channelInfo.models[modelName].output_price,
      maxTokens: channelInfo.models[modelName].max_tokens
    }));

    const tableData = models.map(model => ({
      key: model.model,
      model: model.model,
      inputPrice: formatPrice(model.inputPrice),
      outputPrice: formatPrice(model.outputPrice),
      maxTokens: formatMaxTokens(model.maxTokens)
    }));

    const columns = [
      {
        title: 'Model',
        dataIndex: 'model',
        key: 'model',
        width: '40%'
      },
      {
        title: 'Input Price (per 1M tokens)',
        dataIndex: 'inputPrice',
        key: 'inputPrice',
        width: '25%'
      },
      {
        title: 'Output Price (per 1M tokens)',
        dataIndex: 'outputPrice',
        key: 'outputPrice',
        width: '25%'
      },
      {
        title: 'Max Tokens',
        dataIndex: 'maxTokens',
        key: 'maxTokens',
        width: '10%'
      }
    ];

    return (
      <Card key={channelName} fluid style={{ marginBottom: '1rem' }}>
        <Card.Content>
          <Card.Header>
            {channelName} ({models.length} models)
          </Card.Header>
          <Table celled striped>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell width={6}>Model</Table.HeaderCell>
                <Table.HeaderCell width={3}>Input Price (per 1M tokens)</Table.HeaderCell>
                <Table.HeaderCell width={3}>Output Price (per 1M tokens)</Table.HeaderCell>
                <Table.HeaderCell width={2}>Max Tokens</Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {tableData.map(row => (
                <Table.Row key={row.key}>
                  <Table.Cell>{row.model}</Table.Cell>
                  <Table.Cell>{row.inputPrice}</Table.Cell>
                  <Table.Cell>{row.outputPrice}</Table.Cell>
                  <Table.Cell>{row.maxTokens}</Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        </Card.Content>
      </Card>
    );
  };

  if (loading) {
    return (
      <div className='dashboard-container'>
        <Card fluid className='chart-card'>
          <Card.Content>
            <Loader active inline='centered'>Loading models...</Loader>
          </Card.Content>
        </Card>
      </div>
    );
  }

  const totalModels = Object.values(filteredData).reduce((total, channelInfo) =>
    total + Object.keys(channelInfo.models).length, 0
  );

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            {t('models.title', 'Supported Models')}
          </Card.Header>
          <p>{t('models.description', 'Browse all models supported by the server, grouped by channel/adaptor with pricing information.')}</p>

          <Segment>
            <Grid columns={3} stackable>
              <Grid.Column width={6}>
                <Input
                  icon='search'
                  placeholder={t('models.search_placeholder', 'Search models...')}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  fluid
                />
              </Grid.Column>
              <Grid.Column width={8}>
                <Dropdown
                  placeholder={t('models.filter_channels', 'Filter by channels...')}
                  fluid
                  multiple
                  selection
                  options={Object.keys(modelsData).sort().map(channelName => ({
                    key: channelName,
                    text: `${channelName} (${Object.keys(modelsData[channelName].models).length} models)`,
                    value: channelName
                  }))}
                  value={selectedChannels}
                  onChange={(e, { value }) => setSelectedChannels(value)}
                />
              </Grid.Column>
              <Grid.Column width={2}>
                <Button
                  icon='remove'
                  content={t('models.clear_filters', 'Clear')}
                  onClick={() => {
                    setSearchTerm('');
                    setSelectedChannels([]);
                  }}
                  fluid
                />
              </Grid.Column>
            </Grid>
          </Segment>

          {totalModels === 0 ? (
            <Message info>
              <Message.Header>{t('models.no_results', 'No models found')}</Message.Header>
              <p>{t('models.no_results_description', 'Try adjusting your search terms.')}</p>
            </Message>
          ) : (
            <>
              <Header as='h3'>
                {t('models.results_count', `Found ${totalModels} models in ${Object.keys(filteredData).length} channels`)}
              </Header>
              {Object.keys(filteredData)
                .sort()
                .map(channelName => renderChannelModels(channelName, filteredData[channelName]))}
            </>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default Models;
