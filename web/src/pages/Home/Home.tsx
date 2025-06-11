// import { useNodeStore } from '@/store/nodeStore';
import AutoRestart from '@/components/AutoRestart';
import LogsLoader from '@/components/LogsLoader';
import OnOffBtn from '@/components/OnOffBtn';
import { SelectNetwork } from '@/components/SelectNetwork';
import Intl from '@/i18n/i18n';

export default function Home() {
  return (
      <div className="grid grid-cols-2 gap-5">
        <div className="bg-secondary rounded-2xl w-full max-w-lg p-10">
          <p className="mas-body text-f-primary mb-2">
            {Intl.t('home.title-select-network')}
          </p>
          <SelectNetwork />
        </div>

        <div className="bg-secondary rounded-2xl w-full max-w-lg p-10 flex flex-col justify-center">
          <OnOffBtn />
        </div>

        <div className="bg-secondary rounded-2xl w-full max-w-lg p-10 flex flex-col justify-center">
          <AutoRestart />
        </div>

        <div className="bg-secondary rounded-2xl w-full max-w-lg p-10">
          <LogsLoader />
        </div>
      </div>
  );
}
