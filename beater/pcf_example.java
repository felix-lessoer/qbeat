private void processTarget(MQTarget target) {

             PCFMessageAgent agent = new PCFMessageAgent(getAgentId());

             agent.setWaitInterval(WAITINTERVAL, EXPIRYTIME);

             int processed = 0;

             int excluded = 0;

             int valid = 0;

             try {

                    agent.setAlternateUserId(target.getAlternateuserid());

                    LOGGER.log(Level.INFO, getAgentId() + " AlternateUserId for command execution set to <"+target.getAlternateuserid()+">.");

                    agent.connect(config.getConnection().getHost(), config.getConnection().getPort(), config.getConnection().getChannel(), config.getConnection().getQueuemanager(), target.getTargetqueuemanager(), config.getConnection().getUsername(), config.getConnection().getPassword());

                    LOGGER.log(Level.INFO, getAgentId() + " connected to queuemanager, connection data is <"+config.getConnection().getHost()+":"+config.getConnection().getPort()+">, channel <"+config.getConnection().getChannel()+">. Target queuemanager is <"+target.getTargetqueuemanager()+">.");



                    PCFMessage request = new PCFMessage(CMQCFC.MQCMD_RESET_Q_STATS);

                    LOGGER.log(Level.INFO, getAgentId() + " instantiated <MQCMD_RESET_Q_STATS> request.");

                    request.addParameter(CMQC.MQCA_Q_NAME, target.getQueryobjects());

                    LOGGER.log(Level.INFO, getAgentId() + " added parameter for queues matching to <" + target.getQueryobjects() + ">.");

                    if(agent.getPlatform() == CMQC.MQPL_ZOS && target.getCommandscope() != null) {

                           request.addParameter(CMQCFC.MQCACF_COMMAND_SCOPE, target.getCommandscope());

                           LOGGER.log(Level.INFO, getAgentId() + " Platform Z/OS, added parameter <CMQCFC.MQCACF_COMMAND_SCOPE>, set to <" + target.getCommandscope() + ">.");

                    }

                    Vector<PCFMessage> responses = agent.send(request);

                    LOGGER.log(Level.INFO, getAgentId() + " request sent to <"+target.getTargetqueuemanager()+">. Response contains <" + responses.size() + "> pcf-message(s).");



                    Matcher excludem = Pattern.compile(target.getExcludeobjects()).matcher("");

                    LOGGER.log(Level.INFO, getAgentId() + " instantiated pattern matcher for exclude-queues. Processing response messages.");

                    for (PCFMessage pcfMessage : responses) {

                           processed++;

                           if (pcfMessage.getReason() == CMQC.MQRC_NONE && ( pcfMessage.getType() == CMQCFC.MQCFT_XR_ITEM || pcfMessage.getType() == CMQCFC.MQCFT_RESPONSE ) ) {

                                  if(LOGGER.isEnabledFor(Level.TRACE)) {

                                        agent.printPcfParameters(pcfMessage);

                                  }

                                  String qmgrname;

                                  if(agent.getPlatform() == CMQC.MQPL_ZOS) {

                                        qmgrname = pcfMessage.getStringParameterValue(CMQCFC.MQCACF_RESPONSE_Q_MGR_NAME).trim();

                                  } else {

                                        qmgrname = pcfMessage.getReplyToQueueManagerName().trim();

                                  }

                                  String qname = pcfMessage.getStringParameterValue(CMQC.MQCA_Q_NAME).trim();

                                  if(excludem.reset(qname).matches()) {

                                        getLogger().log(Level.DEBUG, getAgentId() + " Excludepattern match for queue <"+qname+"@"+qmgrname+">. Excluded.");

                                        excluded++;

                                        continue;

                                  }

                                  Integer interval = pcfMessage.getIntParameterValue(CMQC.MQIA_TIME_SINCE_RESET);

                                  if(interval > VALIDINTERVAL) {

                                        getLogger().log(Level.DEBUG, getAgentId() + " Interval <"+interval+"> invalid for queue <"+qname+"@"+qmgrname+">. Excluded.");

                                        excluded++;

                                        continue;

                                  }



                                  JSONObject json = new JSONObject();

                                  json.put("OBJECTTYPE", "MQ_RESET_Q_STATS");

                                  json.put("REPORTTIME", pcfMessage.getPutDateTime().getTimeInMillis());

                                  json.put("MQCA_Q_MGR_NAME", qmgrname);

                                  @SuppressWarnings("unchecked")

                                  Enumeration<PCFParameter> e = pcfMessage.getParameters();

                                  while (e.hasMoreElements()) {

                                        PCFParameter p = (PCFParameter) e.nextElement();

                                        if (p.getType() == CMQCFC.MQCFT_INTEGER) {

                                               json.put(p.getParameterName(), String.valueOf(p.getValue()));

                                        } else {

                                               json.put(p.getParameterName(), p.getStringValue().trim());

                                        }

                                  }

                                  if(getLogger().isEnabledFor(Level.TRACE)) {

                                        getLogger().log(Level.TRACE, getAgentId() + ":\n"+ json.toString(2));

                                  }

                                  evb.fireEvent(new QueueStatisticsEvent<JSONObject>(json));

                                  getLogger().log(Level.DEBUG, getAgentId() + " Queuestatisticsdata for <"+qname+"@"+qmgrname+"> sent to persistence-engine.");

                                  valid++;

                           }      else if (pcfMessage.getReason() == CMQC.MQRC_NONE && pcfMessage.getType() == CMQCFC.MQCFT_XR_SUMMARY) {

                                  if (pcfMessage.getCompCode() == CMQC.MQCC_OK) {

                                        LOGGER.log(Level.INFO, getAgentId() + " processed <MQCMD_RESET_Q_STATS> command successfully on <"+target.getTargetqueuemanager()+">.");

                                  } else if(pcfMessage.getCompCode() == CMQC.MQCC_FAILED) {

                                        LOGGER.log(Level.ERROR, getAgentId() + " Completion code ["+pcfMessage.getCompCode()+"] "+MQConstants.lookupCompCode(pcfMessage.getCompCode())+" with reason [" + pcfMessage.getReason() + "] - "+MQConstants.lookupReasonCode(pcfMessage.getReason())+" occured processing <MQCMD_RESET_Q_STATS> on <"+target.getTargetqueuemanager()+">.");

                                  }

                           }

                    }

                    LOGGER.log(Level.INFO, getAgentId() + " processed data of <" + processed + "> pcf-message(s) from <"+target.getTargetqueuemanager()+">.");

                    LOGGER.log(Level.INFO, getAgentId() + " added data of <" + valid + "> queue(s) from <"+target.getTargetqueuemanager()+">.");

                    LOGGER.log(Level.INFO, getAgentId() + " excluded data of <" + excluded + "> queues.");

             } catch (MQException e) {

                    LOGGER.log(Level.ERROR, getAgentId() + " Completion code ["+e.getCompCode()+"] "+MQConstants.lookupCompCode(e.getCompCode())+" with reason [" + e.getReason() + "] - "+MQConstants.lookupReasonCode(e.getReason())+" occured processing <MQCMD_RESET_Q_STATS> on <"+target.getTargetqueuemanager()+">.");

                    LOGGER.log(Level.TRACE, " Stack: ", e);

             } catch (Exception e) {

                    LOGGER.log(Level.ERROR, getAgentId() + " error retrieving data from <"+target.getTargetqueuemanager()+">: " + e.getMessage());

                    LOGGER.log(Level.TRACE, " Stack: ", e);

             } finally {

                    if (agent != null) {

                           LOGGER.log(Level.INFO, getAgentId() + " Trying to disconnect queuemanager <"+target.getTargetqueuemanager()+">.");

                           try {

                                  agent.disconnect();

                           } catch (Exception e) {

                                  LOGGER.log(Level.ERROR, getAgentId() + " error disconnecting queuemanager <"+target.getTargetqueuemanager()+">: " + e.getMessage());

                                  LOGGER.log(Level.TRACE, " Stack: ", e);

                           }

                           LOGGER.log(Level.INFO, getAgentId() + " disconnected queuemanager <"+target.getTargetqueuemanager()+">.");

                    }

             }

       }
