// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {FormattedMessage} from 'react-intl';
import {useDispatch} from 'react-redux';

import type {Channel} from '@mattermost/types/channels';
import type {UserProfile} from '@mattermost/types/users';

import {isGuest} from 'mattermost-redux/utils/user_utils';

import {openModal} from 'actions/views/modals';

import ConvertGmToChannelModal from 'components/convert_gm_to_channel_modal';
import * as Menu from 'components/menu';

import {ModalIdentifiers} from 'utils/constants';

type Props = {
    isArchived: boolean;
    isReadonly: boolean;
    channel: Channel;
    user: UserProfile;
}

const ConvertGMtoPrivate = ({isArchived, isReadonly, channel, user}: Props): JSX.Element => {
    const dispatch = useDispatch();
    if (isArchived || isReadonly || isGuest(user.roles)) {
        return <></>;
    }

    return (
        <Menu.Item
            id='convertGMPrivateChannel'
            onClick={() => {
                dispatch(
                    openModal({
                        modalId: ModalIdentifiers.CONVERT_GM_TO_CHANNEL,
                        dialogType: ConvertGmToChannelModal,
                        dialogProps: {channel},
                    }),
                );
            }}
            labels={
                <FormattedMessage
                    id='sidebar_left.sidebar_channel_menu_convert_to_channel'
                    defaultMessage='Convert to Private Channel'
                />
            }
        />
    );
};

export default React.memo(ConvertGMtoPrivate);