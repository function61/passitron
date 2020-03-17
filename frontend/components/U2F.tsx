import { DangerAlert, InfoAlert } from 'f61ui/component/alerts';
import { Button } from 'f61ui/component/bootstrap';
import { Loading } from 'f61ui/component/loading';
import { defaultErrorHandler } from 'f61ui/errors';
import { shouldAlwaysSucceed } from 'f61ui/utils';
import { U2FChallengeBundle, U2FResponseBundle } from 'generated/apitypes_types';
import * as React from 'react';
import { isU2FError, nativeSignResultToApiType, u2fErrorMsg, u2fSign } from 'u2ftypes';

interface U2fSignerProps {
	challenge: U2FChallengeBundle;
	signed: (result: U2FResponseBundle) => void;
}

interface U2fSignerState {
	authing: boolean;
	authError?: string;
}

export class U2fSigner extends React.Component<U2fSignerProps, U2fSignerState> {
	state: U2fSignerState = { authing: false };

	componentDidMount() {
		// start fetching process automatically. in some rare cases the user might not
		// want this, but failed auth attempt timeouts are not dangerous and this reduces
		// extra clicks in the majority case
		shouldAlwaysSucceed(this.startSigning());
	}

	render() {
		if (this.state.authing) {
			return (
				<div>
					<InfoAlert>Please swipe your U2F token now ...</InfoAlert>

					<Loading />
				</div>
			);
		}

		return (
			<div>
				<Button
					label="Authenticate"
					click={() => {
						shouldAlwaysSucceed(this.startSigning());
					}}
				/>

				{this.state.authError && <DangerAlert>{this.state.authError}</DangerAlert>}
			</div>
		);
	}

	private async startSigning() {
		this.setState({ authing: true, authError: undefined });

		try {
			// u2fSign should never error/throw
			const result = await u2fSign(this.props.challenge.SignRequest);

			if (isU2FError(result)) {
				this.setState({ authing: false, authError: u2fErrorMsg(result) });
				return;
			}

			this.props.signed({
				Challenge: this.props.challenge.Challenge,
				SignResult: nativeSignResultToApiType(result),
			});
		} catch (e) {
			defaultErrorHandler(e);
		}
	}
}
