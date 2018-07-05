
import { GalleryImage, Gallery, SeedableRNG } from "./gallery";

export class GalleryTileImageLoader extends GalleryImage {

  constructor(url: string, private dim: [number, number], private title: string, private name: string) {
    super(url)
  }

  GetTypeName() : string {
    return "GalleryTileLoader";
  }

  GetDomElement() : HTMLElement {
    let _result = document.createElement('div');
    _result.classList.add('GalleryImage-' + this.GetTypeName());
    _result.classList.add('albumcover');
    _result.onclick = () => { location.href='/gallery/' + this.name };
    _result.style['background-image'] = 'url(' + this.imageUrl + ')';
    _result.style['background-size'] = 'cover';

    let _title = document.createElement('div');
    _title.classList.add('covertitle');
    _title.classList.add('albumletterpress');
    _title.innerText = this.title;

    _result.appendChild(_title)
    return _result;
  }

  GetAspectRatio() : number {
    return this.dim[0] / this.dim[1]
  }
}

export class PreSizedFadeInImageLoader extends GalleryImage {

  constructor(url: string, private dimensions: [number, number]) {
    super(url);
  }

  GetTypeName(): string {
    return "PreSizedFadeInImageLoader"
  }

  GetDomElement(): HTMLElement {
    let _dom_element = document.createElement('div');
    _dom_element.style.animation = 'fadein 5s';
    _dom_element.style['background-size'] = 'cover';
    _dom_element.classList.add('GalleryImage-'+this.GetTypeName());

    let _image_preload = new Image();
    _image_preload.onload = function() {
      _dom_element.style['background-image'] = "url("+_image_preload.src+")";
    }
    _image_preload.src = this.imageUrl;

    return _dom_element;
  }

  GetAspectRatio(): number {
    return this.dimensions[0] / this.dimensions[1];
  }

}
