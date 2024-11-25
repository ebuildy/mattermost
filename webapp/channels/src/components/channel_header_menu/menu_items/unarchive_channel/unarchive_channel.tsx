// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {memo} from 'react';
import {FormattedMessage} from 'react-intl';
import {useDispatch} from 'react-redux';

import type {Channel} from '@mattermost/types/channels';

import {openModal} from 'actions/views/modals';

import * as Menu from 'components/menu';
import UnarchiveChannelModal from 'components/unarchive_channel_modal';

import {ModalIdentifiers} from 'utils/constants';

type Props = {
    channel: Channel;
    isDefault: boolean;
    isArchived: boolean;
}

const UnarchiveChannel = ({
    isDefault = true,
    isArchived = false,
    channel,
}: Props) => {
    const dispatch = useDispatch();

    if (!isArchived || isDefault) {
        return <></>;
    }

    return (
        <>
            <Menu.Separator/>
            <Menu.Item
                id='channelUnarchiveChannel'
                onClick={() => {
                    dispatch(
                        openModal({
                            modalId: ModalIdentifiers.UNARCHIVE_CHANNEL,
                            dialogType: UnarchiveChannelModal,
                            dialogProps: {channel},
                        }),
                    );
                }}
                labels={
                    <FormattedMessage
                        id='channel_header.unarchive'
                        defaultMessage='Unarchive Channel'
                    />
                }
            />
        </>
    );
};

export default memo(UnarchiveChannel);